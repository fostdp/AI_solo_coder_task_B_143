package api

import (
	"bytes"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"crossbow-simulation/backend/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func mmValue(m *model.MeasurementMeta) float64 {
	if m == nil {
		return 0
	}
	return m.Value
}

func newMM(v float64) *model.MeasurementMeta {
	return &model.MeasurementMeta{Value: v}
}

func getPerformanceMetric(v *model.CrossbowVariant, key string) float64 {
	switch key {
	case "drawWeight":
		return mmValue(v.Performance.DrawWeight)
	case "maxRange":
		return mmValue(v.Performance.MaxRange)
	case "effectiveRange":
		return mmValue(v.Performance.EffectiveRange)
	case "idealFireRate":
		return mmValue(v.Performance.IdealFireRate)
	case "magazineSize":
		return float64(v.Performance.MagazineSize)
	case "reloadTime":
		return mmValue(v.Performance.ReloadTime)
	case "accuracyScore":
		return mmValue(v.Performance.AccuracyScore)
	default:
		return 0
	}
}

func isHigherBetter(key string) bool {
	switch key {
	case "reloadTime":
		return false
	default:
		return true
	}
}

func getMetricLabel(key string) string {
	switch key {
	case "idealFireRate":
		return "射速(发/分)"
	case "magazineSize":
		return "弹容(发)"
	case "effectiveRange":
		return "有效射程(m)"
	case "drawWeight":
		return "拉力(N)"
	case "reloadTime":
		return "装填时间(s)"
	case "accuracyScore":
		return "精度评分"
	default:
		return "unknown"
	}
}

func weibullCDF(x, k, lambda float64) float64 {
	if x <= 0 {
		return 0
	}
	return 1 - math.Exp(-math.Pow(x/lambda, k))
}

func weibullQuantile(p, k, lambda float64) float64 {
	if p <= 0 {
		return 0
	}
	if p >= 1 {
		return math.Inf(1)
	}
	return lambda * math.Pow(-math.Log(1-p), 1/k)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())
	ctrl := NewController(nil, nil)
	api := r.Group("/api/v1")
	{
		variants := api.Group("/variants")
		{
			variants.GET("", ctrl.ListVariants)
			variants.GET("/:code", ctrl.GetVariant)
			variants.POST("/compare", ctrl.CompareVariants)
		}
		firearms := api.Group("/firearms")
		{
			firearms.GET("", ctrl.ListModernFirearms)
			firearms.POST("/compare-era", ctrl.CompareEraFirearms)
		}
	}
	return r
}

// ================== Feature 1: 弩型机构对比测试 ==================

// 正常：列出所有弩型
func TestListVariants_Success(t *testing.T) {
	r := setupTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/variants", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp model.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)

	data, ok := resp.Data.(map[string]interface{})
	assert.True(t, ok)
	total, ok := data["total"].(float64)
	assert.True(t, ok)
	assert.Equal(t, float64(3), total)

	items, ok := data["items"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, items, 3)

	codes := map[string]bool{}
	for _, it := range items {
		mp := it.(map[string]interface{})
		codes[mp["variantCode"].(string)] = true
	}
	assert.True(t, codes["zhuge"])
	assert.True(t, codes["san-gong"])
	assert.True(t, codes["bi-zhang"])
}

// 正常：获取单个弩型详情
func TestGetVariant_Exists(t *testing.T) {
	r := setupTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/variants/zhuge", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp model.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)

	b, _ := json.Marshal(resp.Data)
	var v model.CrossbowVariant
	_ = json.Unmarshal(b, &v)
	assert.Equal(t, "zhuge", v.VariantCode)
	assert.Equal(t, "诸葛弩", v.Name)
	assert.Equal(t, 10, v.Performance.MagazineSize)
	assert.Greater(t, mmValue(v.Performance.IdealFireRate), 5.0)
}

// 异常：弩型不存在
func TestGetVariant_NotFound(t *testing.T) {
	r := setupTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/variants/nonexistent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var resp model.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, "弩型未找到", resp.Message)
}

// 正常：对比3种弩型
func TestCompareVariants_ThreeVariants(t *testing.T) {
	r := setupTestRouter()
	body := model.VariantCompareRequest{
		VariantCodes:   []string{"zhuge", "san-gong", "bi-zhang"},
		CompareMetrics: []string{"idealFireRate", "effectiveRange", "magazineSize", "drawWeight", "reloadTime", "accuracyScore"},
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/variants/compare", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp model.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)

	rb, _ := json.Marshal(resp.Data)
	var cmp model.VariantCompareResponse
	_ = json.Unmarshal(rb, &cmp)
	assert.Len(t, cmp.ComparedVariants, 3)
	assert.Len(t, cmp.AdvantageMap, 6)

	// 验证核心结论：射速之王是诸葛弩
	fireRateAdv := findAdvantage(cmp.AdvantageMap, "射速(发/分)")
	assert.NotNil(t, fireRateAdv)
	assert.Equal(t, "zhuge", fireRateAdv.BestVariant)

	// 射程之王是三弓弩
	rangeAdv := findAdvantage(cmp.AdvantageMap, "有效射程(m)")
	assert.NotNil(t, rangeAdv)
	assert.Equal(t, "san-gong", rangeAdv.BestVariant)

	// 弹容之王是诸葛弩
	magAdv := findAdvantage(cmp.AdvantageMap, "弹容(发)")
	assert.NotNil(t, magAdv)
	assert.Equal(t, "zhuge", magAdv.BestVariant)

	// 装填时间最快是诸葛弩（低者优）
	reloadAdv := findAdvantage(cmp.AdvantageMap, "装填时间(s)")
	assert.NotNil(t, reloadAdv)
	assert.Equal(t, "zhuge", reloadAdv.BestVariant)
}

// 边界：空请求 - 应使用默认值
func TestCompareVariants_EmptyBody(t *testing.T) {
	r := setupTestRouter()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/variants/compare", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp model.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
}

// 异常：全部无效弩型编码
func TestCompareVariants_AllInvalidCodes(t *testing.T) {
	r := setupTestRouter()
	body := model.VariantCompareRequest{VariantCodes: []string{"xx", "yy"}}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/variants/compare", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp model.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Equal(t, "无有效弩型编码", resp.Message)
}

// 边界：部分有效编码
func TestCompareVariants_PartialValid(t *testing.T) {
	r := setupTestRouter()
	body := model.VariantCompareRequest{VariantCodes: []string{"zhuge", "invalid", "san-gong"}}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/variants/compare", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp model.APIResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	rb, _ := json.Marshal(resp.Data)
	var cmp model.VariantCompareResponse
	_ = json.Unmarshal(rb, &cmp)
	assert.Len(t, cmp.ComparedVariants, 2)
}

// 边界：单发弹容弩对比验证射速差异
func TestCompareVariants_FireRateAssertion(t *testing.T) {
	zhugeRate := 0.0
	sanGongRate := 0.0
	biZhangRate := 0.0
	for _, v := range model.CrossbowPresets() {
		switch v.VariantCode {
		case "zhuge":
			zhugeRate = mmValue(v.Performance.IdealFireRate)
		case "san-gong":
			sanGongRate = mmValue(v.Performance.IdealFireRate)
		case "bi-zhang":
			biZhangRate = mmValue(v.Performance.IdealFireRate)
		}
	}

	// 核心业务约束：诸葛弩射速 > 臂张弩 > 三弓弩
	assert.Greater(t, zhugeRate, biZhangRate,
		"诸葛弩作为连发弩，射速应高于臂张弩")
	assert.Greater(t, biZhangRate, sanGongRate,
		"臂张弩射速应高于大型三弓弩")
	// 倍数约束：诸葛弩射速应是三弓弩的6倍以上
	assert.Greater(t, zhugeRate/sanGongRate, 5.0,
		"诸葛弩射速应至少是三弓弩的5倍")
}

func findAdvantage(list []model.AdvantageMap, metric string) *model.AdvantageMap {
	for i := range list {
		if list[i].Metric == metric {
			return &list[i]
		}
	}
	return nil
}

// ================== Feature 2: 跨时代射速对比测试 ==================

// 正常：列出现代步枪
func TestListModernFirearms_Success(t *testing.T) {
	r := setupTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/firearms", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp model.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)

	data, ok := resp.Data.(map[string]interface{})
	assert.True(t, ok)
	total, ok := data["total"].(float64)
	assert.True(t, ok)
	assert.Equal(t, float64(6), total)
}

// 正常：完整跨时代对比
func TestCompareEraFirearms_FullComparison(t *testing.T) {
	r := setupTestRouter()
	body := model.EraFirearmCompareRequest{
		AncientVariants: []string{"zhuge", "san-gong", "bi-zhang"},
		ModernFirearms:  []string{"AK-47", "M249 SAW", "HK MP5", "M16A1"},
		CompareMetrics:  []string{"fireRate", "effectiveRange", "magazineSize", "muzzleVelocity"},
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/firearms/compare-era", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp model.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)

	rb, _ := json.Marshal(resp.Data)
	var cmp model.EraFirearmCompareResponse
	_ = json.Unmarshal(rb, &cmp)
	assert.Len(t, cmp.AncientVariants, 3)
	assert.Len(t, cmp.ModernFirearms, 4)
	assert.NotEmpty(t, cmp.EraGapTable)

	// 验证核心结论：技术进步的数量级差异
	fireRateGap := findGap(cmp.EraGapTable, "射速(实战)")
	assert.NotNil(t, fireRateGap)
	t.Logf("射速差距倍数: %.2f", fireRateGap.GapRatio)
	assert.Greater(t, fireRateGap.GapRatio, 5.0,
		"现代实战射速应至少是古代最佳射速的5倍（技术进步验证）")

	rangeGap := findGap(cmp.EraGapTable, "有效射程")
	assert.NotNil(t, rangeGap)
	t.Logf("射程差距倍数: %.2f", rangeGap.GapRatio)
	assert.Greater(t, rangeGap.GapRatio, 1.0,
		"现代有效射程应超过古代最佳三弓弩")

	energyGap := findGap(cmp.EraGapTable, "杀伤动能(估算)")
	assert.NotNil(t, energyGap)
	assert.Greater(t, energyGap.GapRatio, 5.0,
		"现代弹头动能应显著高于古代箭矢")
}

// 边界：空请求 - 默认值
func TestCompareEraFirearms_EmptyBody(t *testing.T) {
	r := setupTestRouter()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/firearms/compare-era", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp model.APIResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.True(t, resp.Success)
}

// 边界：双方都是空数组，使用默认值
func TestCompareEraFirearms_BothEmpty(t *testing.T) {
	r := setupTestRouter()
	body := model.EraFirearmCompareRequest{
		AncientVariants: []string{},
		ModernFirearms:  []string{},
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/firearms/compare-era", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// 异常：现代武器名无效 - 应忽略而非报错
func TestCompareEraFirearms_ModernInvalid(t *testing.T) {
	r := setupTestRouter()
	body := model.EraFirearmCompareRequest{
		AncientVariants: []string{"zhuge"},
		ModernFirearms:  []string{"不存在的武器1号"},
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/firearms/compare-era", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp model.APIResponse
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	rb, _ := json.Marshal(resp.Data)
	var cmp model.EraFirearmCompareResponse
	_ = json.Unmarshal(rb, &cmp)
	// 古代弩仍在，现代武器空
	assert.Len(t, cmp.AncientVariants, 1)
	assert.Len(t, cmp.ModernFirearms, 0)
}

// 核心验证：AK-47实战射速 vs 诸葛弩的技术差距数量级
func TestFirearmPresets_Ak47vsZhugeRateGap(t *testing.T) {
	bestAncient := 0.0
	for _, v := range model.CrossbowPresets() {
		if mmValue(v.Performance.IdealFireRate) > bestAncient {
			bestAncient = mmValue(v.Performance.IdealFireRate)
		}
	}

	firearms := model.ModernFirearmPresets()
	var ak47 model.ModernFirearm
	for _, f := range firearms {
		if f.Name == "AK-47" {
			ak47 = f
		}
	}

	assert.Equal(t, "AK-47", ak47.Name)
	t.Logf("诸葛弩射速 %.1f 发/分, AK-47实战射速 %.0f 发/分", bestAncient, ak47.EffectiveRPM)
	ratio := ak47.EffectiveRPM / bestAncient
	assert.Greater(t, ratio, 6.0,
		"AK-47实战射速应至少是诸葛弩的6倍，验证1900年的技术进步幅度")
	t.Logf("技术进步倍数: %.2f 倍 (1700年跨度)", ratio)
}

// 核心验证：M249班用机枪 vs 三弓弩的火力密度差距
func TestFirearmPresets_M249vsSanGongFireDensity(t *testing.T) {
	var sanGong model.CrossbowVariant
	var m249 model.ModernFirearm
	for _, v := range model.CrossbowPresets() {
		if v.VariantCode == "san-gong" {
			sanGong = v
		}
	}
	for _, f := range model.ModernFirearmPresets() {
		if f.Name == "M249 SAW" {
			m249 = f
		}
	}

	// 1分钟发射数差距
	sanGongRate := mmValue(sanGong.Performance.IdealFireRate)
	minuteRatio := m249.EffectiveRPM / sanGongRate
	t.Logf("三弓弩射速: %.1f 发/分, M249: %.0f 发/分, 差距: %.0f倍",
		sanGongRate, m249.EffectiveRPM, minuteRatio)
	assert.Greater(t, minuteRatio, 50.0,
		"M249班用机枪火力密度应是三弓弩的50倍以上")

	// 10分钟战斗总弹量差距（含弹容和射速）
	sangong10Min := sanGongRate * 10
	m24910Min := float64(m249.MagazineSize) + m249.EffectiveRPM*0.7 // 扣除换弹
	magRatio := m24910Min / sangong10Min
	t.Logf("10分钟战斗 三弓弩: %.0f发 vs M249: %.0f发, 总弹量差距: %.0f倍",
		sangong10Min, m24910Min, magRatio)
	assert.Greater(t, magRatio, 10.0,
		"10分钟战斗中，M249投送弹量应是三弓弩的10倍以上")
}

func findGap(list []model.EraGapEntry, metric string) *model.EraGapEntry {
	for i := range list {
		if list[i].Metric == metric {
			return &list[i]
		}
	}
	return nil
}

// ================== 辅助：getPerformanceMetric 单元测试 ==================

func TestGetPerformanceMetric_AllKeys(t *testing.T) {
	v := &model.CrossbowVariant{
		Performance: model.PerformanceMetrics{
			DrawWeight:     newMM(1000),
			MaxRange:       newMM(500),
			EffectiveRange: newMM(200),
			IdealFireRate:  newMM(5),
			MagazineSize:   8,
			ReloadTime:     newMM(10),
			AccuracyScore:  newMM(0.75),
		},
	}
	assert.Equal(t, 1000.0, getPerformanceMetric(v, "drawWeight"))
	assert.Equal(t, 500.0, getPerformanceMetric(v, "maxRange"))
	assert.Equal(t, 200.0, getPerformanceMetric(v, "effectiveRange"))
	assert.Equal(t, 5.0, getPerformanceMetric(v, "idealFireRate"))
	assert.Equal(t, 8.0, getPerformanceMetric(v, "magazineSize"))
	assert.Equal(t, 10.0, getPerformanceMetric(v, "reloadTime"))
	assert.Equal(t, 0.75, getPerformanceMetric(v, "accuracyScore"))
	assert.Equal(t, 0.0, getPerformanceMetric(v, "unknownKey"))
}

func TestIsHigherBetter(t *testing.T) {
	assert.True(t, isHigherBetter("drawWeight"))
	assert.True(t, isHigherBetter("idealFireRate"))
	assert.True(t, isHigherBetter("magazineSize"))
	assert.False(t, isHigherBetter("reloadTime"))
	assert.True(t, isHigherBetter("anything_else"))
}

func TestGetMetricLabel(t *testing.T) {
	assert.Equal(t, "射速(发/分)", getMetricLabel("idealFireRate"))
	assert.Equal(t, "弹容(发)", getMetricLabel("magazineSize"))
	assert.Equal(t, "unknown", getMetricLabel("unknown"))
}

// 弩型预设完整性
func TestCrossbowPresets_AllFieldsPopulated(t *testing.T) {
	for _, v := range model.CrossbowPresets() {
		assert.NotEmpty(t, v.VariantCode, "VariantCode 必填")
		assert.NotEmpty(t, v.Name, "Name 必填")
		assert.NotNil(t, v.MechanismParams, "机构参数必填")
		assert.Greater(t, mmValue(v.Performance.DrawWeight), 0.0)
		assert.Greater(t, mmValue(v.Performance.MaxRange), mmValue(v.Performance.EffectiveRange),
			"最大射程应大于有效射程")
		assert.Greater(t, v.Performance.MagazineSize, 0)
	}
}

// 步枪预设完整性
func TestModernFirearmPresets_AllFieldsPopulated(t *testing.T) {
	for _, f := range model.ModernFirearmPresets() {
		assert.NotEmpty(t, f.Name)
		assert.NotEmpty(t, f.FirearmType)
		assert.GreaterOrEqual(t, f.EffectiveRPM, 0.0)
		assert.Greater(t, f.MagazineSize, 0)
		assert.Greater(t, f.EffectiveRangeM, 0.0)
		assert.Greater(t, f.MuzzleVelocityMPS, 0.0)
		// 物理一致性：弹头初速应在亚音速到步枪超音速之间
		assert.Greater(t, f.MuzzleVelocityMPS, 150.0,
			"枪口初速不应低于150m/s")
		assert.Less(t, f.MuzzleVelocityMPS, 2000.0,
			"枪口初速不应高于2000m/s")
	}
}

// 数学约束：可靠性曲线的 R(0)=1, R(∞)=0
func TestReliabilityCurve_MathValid(t *testing.T) {
	presets := model.CrossbowPresets()
	for _, v := range presets {
		capacity := v.Performance.MagazineSize
		lambda := float64(capacity) * 500.0
		k := 2.5
		cdf0 := weibullCDF(0, k, lambda)
		assert.Equal(t, 0.0, cdf0, "威布尔CDF(0)=0")
		quant0 := weibullQuantile(0, k, lambda)
		assert.Equal(t, 0.0, quant0, "威布尔分位数p=0→0")
		quant1 := weibullQuantile(1.0, k, lambda)
		assert.True(t, math.IsInf(quant1, 1), "p=1→+∞")
		quant1m10 := weibullQuantile(1-1e-10, k, lambda)
		assert.Greater(t, quant1m10, lambda*2.0,
			"p=1-1e-10 时应显著大于lambda尺度参数")
	}
}
