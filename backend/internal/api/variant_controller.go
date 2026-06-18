package api

import (
	"math"
	"net/http"
	"sort"
	"strconv"

	"crossbow-simulation/backend/internal/model"
	"crossbow-simulation/backend/internal/reliability"
	"crossbow-simulation/backend/internal/virtual_shoot"

	"github.com/gin-gonic/gin"
)

func buildVariantMap() map[string]model.CrossbowVariant {
	m := make(map[string]model.CrossbowVariant)
	for _, v := range model.CrossbowPresets() {
		m[v.VariantCode] = v
	}
	return m
}

func buildFirearmMap() map[string]model.ModernFirearm {
	m := make(map[string]model.ModernFirearm)
	for _, f := range model.ModernFirearmPresets() {
		m[f.Name] = f
	}
	return m
}

func getPerformanceMetric(v *model.CrossbowVariant, metric string) float64 {
	switch metric {
	case "drawWeight":
		return v.Performance.DrawWeight
	case "maxRange":
		return v.Performance.MaxRange
	case "effectiveRange":
		return v.Performance.EffectiveRange
	case "idealFireRate":
		return v.Performance.IdealFireRate
	case "magazineSize":
		return float64(v.Performance.MagazineSize)
	case "reloadTime":
		return v.Performance.ReloadTime
	case "accuracyScore":
		return v.Performance.AccuracyScore
	default:
		return 0
	}
}

func isHigherBetter(metric string) bool {
	switch metric {
	case "reloadTime":
		return false
	default:
		return true
	}
}

func getMetricLabel(metric string) string {
	switch metric {
	case "drawWeight":
		return "弩臂张力(N)"
	case "maxRange":
		return "最大射程(m)"
	case "effectiveRange":
		return "有效射程(m)"
	case "idealFireRate":
		return "射速(发/分)"
	case "magazineSize":
		return "弹容(发)"
	case "reloadTime":
		return "装填时间(s)"
	case "accuracyScore":
		return "精度评分(0-1)"
	default:
		return metric
	}
}

func (ctrl *Controller) ListVariants(c *gin.Context) {
	presets := model.CrossbowPresets()
	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"items": presets,
			"total": len(presets),
		},
	})
}

func (ctrl *Controller) GetVariant(c *gin.Context) {
	code := c.Param("code")
	vm := buildVariantMap()
	v, ok := vm[code]
	if !ok {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Message: "弩型未找到",
		})
		return
	}
	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data:    v,
	})
}

func (ctrl *Controller) CompareVariants(c *gin.Context) {
	var req model.VariantCompareRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.VariantCodes) == 0 {
		req.VariantCodes = []string{"zhuge", "san-gong", "bi-zhang"}
	}
	if len(req.CompareMetrics) == 0 {
		req.CompareMetrics = []string{"drawWeight", "maxRange", "effectiveRange", "idealFireRate", "magazineSize", "reloadTime", "accuracyScore"}
	}

	vm := buildVariantMap()
	compared := make([]model.CrossbowVariant, 0, len(req.VariantCodes))
	codeList := make([]string, 0)
	for _, code := range req.VariantCodes {
		if v, ok := vm[code]; ok {
			compared = append(compared, v)
			codeList = append(codeList, code)
		}
	}
	if len(compared) == 0 {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "无有效弩型编码",
		})
		return
	}

	radar := make([]model.PerformanceRadar, 0, len(req.CompareMetrics))
	advantages := make([]model.AdvantageMap, 0, len(req.CompareMetrics))

	for _, metric := range req.CompareMetrics {
		vals := make(map[string]float64)
		higher := isHigherBetter(metric)
		type kv struct {
			Code  string
			Value float64
		}
		var ranked []kv
		for i, v := range compared {
			val := getPerformanceMetric(&v, metric)
			vals[codeList[i]] = val
			ranked = append(ranked, kv{Code: codeList[i], Value: val})
		}
		sort.Slice(ranked, func(i, j int) bool {
			if higher {
				return ranked[i].Value > ranked[j].Value
			}
			return ranked[i].Value < ranked[j].Value
		})

		best := ranked[0]
		runnerUpCode := ""
		ratio := 0.0
		if len(ranked) >= 2 {
			ru := ranked[1]
			runnerUpCode = ru.Code
			if best.Value != 0 {
				if higher {
					ratio = best.Value / math.Max(0.0001, ru.Value)
				} else {
					ratio = ru.Value / math.Max(0.0001, best.Value)
				}
			}
		}

		radar = append(radar, model.PerformanceRadar{
			Metric: getMetricLabel(metric),
			Values: vals,
			Best:   best.Code,
		})
		advantages = append(advantages, model.AdvantageMap{
			Metric:         getMetricLabel(metric),
			BestVariant:    best.Code,
			BestValue:      best.Value,
			RunnerUp:       runnerUpCode,
			AdvantageRatio: ratio,
		})
	}

	resp := model.VariantCompareResponse{
		ComparedVariants: compared,
		PerformanceRadar: radar,
		AdvantageMap:     advantages,
	}
	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data:    resp,
	})
}

func (ctrl *Controller) ListModernFirearms(c *gin.Context) {
	firearms := model.ModernFirearmPresets()
	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"items": firearms,
			"total": len(firearms),
		},
	})
}

func (ctrl *Controller) CompareEraFirearms(c *gin.Context) {
	var req model.EraFirearmCompareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req = model.EraFirearmCompareRequest{
			AncientVariants: []string{"zhuge", "san-gong", "bi-zhang"},
			ModernFirearms:  []string{"AK-47", "M249 SAW"},
		}
	}
	if len(req.AncientVariants) == 0 {
		req.AncientVariants = []string{"zhuge", "san-gong", "bi-zhang"}
	}
	if len(req.ModernFirearms) == 0 {
		req.ModernFirearms = []string{"AK-47", "M249 SAW"}
	}
	if len(req.CompareMetrics) == 0 {
		req.CompareMetrics = []string{"fireRate", "effectiveRange", "magazineSize", "muzzleVelocity"}
	}

	vm := buildVariantMap()
	fm := buildFirearmMap()

	ancientList := make([]model.CrossbowVariant, 0)
	for _, c := range req.AncientVariants {
		if v, ok := vm[c]; ok {
			ancientList = append(ancientList, v)
		}
	}
	modernList := make([]model.ModernFirearm, 0)
	for _, n := range req.ModernFirearms {
		if f, ok := fm[n]; ok {
			modernList = append(modernList, f)
		}
	}

	bestAncientFireRate := 0.0
	bestAncientRange := 0.0
	bestAncientMag := 0
	for _, v := range ancientList {
		if v.Performance.IdealFireRate > bestAncientFireRate {
			bestAncientFireRate = v.Performance.IdealFireRate
		}
		if v.Performance.EffectiveRange > bestAncientRange {
			bestAncientRange = v.Performance.EffectiveRange
		}
		if v.Performance.MagazineSize > bestAncientMag {
			bestAncientMag = v.Performance.MagazineSize
		}
	}

	bestModernFireRate := 0.0
	bestModernRange := 0.0
	bestModernMag := 0
	bestModernMV := 0.0
	for _, f := range modernList {
		fr := f.EffectiveRPM
		if fr == 0 {
			fr = f.CyclicRateRPM * 0.15
		}
		if fr > bestModernFireRate {
			bestModernFireRate = fr
		}
		if f.EffectiveRangeM > bestModernRange {
			bestModernRange = f.EffectiveRangeM
		}
		if f.MagazineSize > bestModernMag {
			bestModernMag = f.MagazineSize
		}
		if f.MuzzleVelocityMPS > bestModernMV {
			bestModernMV = f.MuzzleVelocityMPS
		}
	}

	if bestAncientFireRate <= 0 {
		bestAncientFireRate = 10
	}
	if bestAncientRange <= 0 {
		bestAncientRange = 300
	}
	if bestAncientMag == 0 {
		bestAncientMag = 10
	}
	if bestModernFireRate <= 0 {
		bestModernFireRate = 100
	}
	if bestModernRange <= 0 {
		bestModernRange = 400
	}
	if bestModernMag == 0 {
		bestModernMag = 30
	}
	if bestModernMV <= 0 {
		bestModernMV = 700
	}

	gapTable := []model.EraGapEntry{
		{
			Metric:       "射速(实战)",
			AncientValue: bestAncientFireRate,
			AncientUnit:  "发/分",
			ModernValue:  bestModernFireRate,
			ModernUnit:   "发/分",
			GapRatio:     bestModernFireRate / bestAncientFireRate,
			Remark:       "自动武器的循环射速使火力密度提升数量级",
		},
		{
			Metric:       "有效射程",
			AncientValue: bestAncientRange,
			AncientUnit:  "米",
			ModernValue:  bestModernRange,
			ModernUnit:   "米",
			GapRatio:     bestModernRange / bestAncientRange,
			Remark:       "线膛枪管与定装弹显著延伸有效射程",
		},
		{
			Metric:       "弹容",
			AncientValue: float64(bestAncientMag),
			AncientUnit:  "发",
			ModernValue:  float64(bestModernMag),
			ModernUnit:   "发",
			GapRatio:     float64(bestModernMag) / math.Max(1, float64(bestAncientMag)),
			Remark:       "盒式弹匣与弹链使持续火力大幅提高",
		},
		{
			Metric:       "初速",
			AncientValue: 80,
			AncientUnit:  "m/s",
			ModernValue:  bestModernMV,
			ModernUnit:   "m/s",
			GapRatio:     bestModernMV / 80.0,
			Remark:       "火药燃气提供的能量远超人力张弩",
		},
		{
			Metric:       "杀伤动能(估算)",
			AncientValue: 160,
			AncientUnit:  "J",
			ModernValue:  2000,
			ModernUnit:   "J",
			GapRatio:     2000.0 / 160.0,
			Remark:       "基于箭重50g@80m/s vs 7.62mm弹头8g@715m/s估算",
		},
	}

	resp := model.EraFirearmCompareResponse{
		AncientVariants: ancientList,
		ModernFirearms:  modernList,
		EraGapTable:     gapTable,
	}
	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data:    resp,
	})
}

func (ctrl *Controller) AnalyzeMagazineReliability(c *gin.Context) {
	code := c.Param("code")
	vm := buildVariantMap()
	variant, ok := vm[code]
	if !ok {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Message: "弩型未找到",
		})
		return
	}

	var body struct {
		Shots     int     `json:"shots"`
		SimTimeSec float64 `json:"simTimeSec"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		body.Shots, _ = strconv.Atoi(c.DefaultQuery("shots", "5000"))
		st, _ := strconv.ParseFloat(c.DefaultQuery("simTime", "7200"), 64)
		body.SimTimeSec = st
	}
	if body.Shots <= 0 {
		body.Shots = 5000
	}
	if body.SimTimeSec <= 0 {
		body.SimTimeSec = 7200
	}

	params := reliability.BuildParamsFromVariant(&variant)
	analyzer := reliability.NewMagazineReliabilityAnalyzer(params)
	analysis := analyzer.Analyze(body.Shots, body.SimTimeSec)

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"variant":  variant,
			"analysis": analysis,
		},
	})
}

func (ctrl *Controller) StartVirtualShoot(c *gin.Context) {
	var req struct {
		VariantCode string `json:"variantCode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.VariantCode = c.DefaultQuery("variant", "zhuge")
	}
	if req.VariantCode == "" {
		req.VariantCode = "zhuge"
	}

	if ctrl.shootMgr == nil {
		ctrl.shootMgr = virtual_shoot.NewVirtualShootManager()
	}

	sess, err := ctrl.shootMgr.NewSession(req.VariantCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Message: "创建射击会话失败",
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data:    sess,
	})
}

func (ctrl *Controller) ShootAction(c *gin.Context) {
	var req model.VirtualShootRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "无效的请求参数",
		})
		return
	}
	if req.SessionID == "" {
		c.JSON(http.StatusBadRequest, model.APIResponse{
			Success: false,
			Message: "缺少 sessionId",
		})
		return
	}

	if ctrl.shootMgr == nil {
		ctrl.shootMgr = virtual_shoot.NewVirtualShootManager()
	}

	resp, err := ctrl.shootMgr.Shoot(req.SessionID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.APIResponse{
			Success: false,
			Message: "射击失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data:    resp,
	})
}

func (ctrl *Controller) GetVirtualShootStatus(c *gin.Context) {
	sessionID := c.Param("id")
	if ctrl.shootMgr == nil {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Message: "会话不存在",
		})
		return
	}
	sess, ok := ctrl.shootMgr.GetSession(sessionID)
	if !ok {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Message: "会话不存在",
		})
		return
	}
	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Data:    sess,
	})
}

func (ctrl *Controller) ResetVirtualShoot(c *gin.Context) {
	sessionID := c.Param("id")
	if ctrl.shootMgr == nil {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Message: "会话不存在",
		})
		return
	}
	sess, ok := ctrl.shootMgr.ResetSession(sessionID)
	if !ok {
		c.JSON(http.StatusNotFound, model.APIResponse{
			Success: false,
			Message: "会话不存在",
		})
		return
	}
	c.JSON(http.StatusOK, model.APIResponse{
		Success: true,
		Message: "会话已重置",
		Data:    sess,
	})
}
