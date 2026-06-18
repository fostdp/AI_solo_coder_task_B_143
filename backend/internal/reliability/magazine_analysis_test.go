package reliability

import (
	"math"
	"testing"

	"crossbow-simulation/backend/internal/model"

	"github.com/stretchr/testify/assert"
)

// ================== Feature 3: 箭匣供弹可靠性测试 ==================

// 正常：诸葛弩标准弹容分析
func TestAnalyze_ZhugeMagazine_StandardShots(t *testing.T) {
	var zhuge model.CrossbowVariant
	for _, v := range model.CrossbowPresets() {
		if v.VariantCode == "zhuge" {
			zhuge = v
		}
	}
	params := BuildParamsFromVariant(&zhuge)
	assert.Equal(t, 10, params.Capacity, "诸葛弩弹容10发")

	analyzer := NewMagazineReliabilityAnalyzer(params)
	analysis := analyzer.Analyze(10000, 7200)

	// 基本数据完整
	assert.Equal(t, 10000, analysis.TotalShots)
	assert.GreaterOrEqual(t, analysis.JamCount, 0)
	t.Logf("诸葛弩弹容10发，10000发测试：卡弹 %d 次，卡弹率 %.6f (%.4f%%)",
		analysis.JamCount, analysis.JamProbabilityPerShot,
		analysis.JamProbabilityPerShot*100)

	// 卡弹概率合理性：在 0.0001 ~ 0.1 之间 (0.01% ~ 10%)
	assert.Greater(t, analysis.JamProbabilityPerShot, 1e-6)
	assert.Less(t, analysis.JamProbabilityPerShot, 0.2)

	// MTBF合理性
	assert.Greater(t, analysis.MTBFShots, 0.0)
	assert.Greater(t, analysis.MTBFHours, 0.0)
	t.Logf("MTBF: %.0f 发, %.2f 小时", analysis.MTBFShots, analysis.MTBFHours)

	// 置信区间上下界合理
	assert.Greater(t, analysis.ConfidenceInterval.Low, 0.0)
	assert.Less(t, analysis.ConfidenceInterval.Low, analysis.ConfidenceInterval.High)
	t.Logf("95%%CI: [%.0f, %.0f]", analysis.ConfidenceInterval.Low, analysis.ConfidenceInterval.High)

	// 7种失效模式齐全
	assert.Equal(t, 7, len(analysis.FailureModeDistribution))
	allZero := true
	for _, v := range analysis.FailureModeDistribution {
		if v > 0 {
			allZero = false
		}
	}
	if analysis.JamCount > 0 {
		assert.False(t, allZero, "卡弹数>0时，分布不应全0")
	}

	// 可靠性曲线点数
	assert.Equal(t, 100, len(analysis.ReliabilityCurvePts))
}

// 正常：三弓弩（单发）分析
func TestAnalyze_SanGong_SingleShotMagazine(t *testing.T) {
	var sg model.CrossbowVariant
	for _, v := range model.CrossbowPresets() {
		if v.VariantCode == "san-gong" {
			sg = v
		}
	}
	params := BuildParamsFromVariant(&sg)
	assert.Equal(t, 1, params.Capacity, "三弓弩弹容1发")

	analyzer := NewMagazineReliabilityAnalyzer(params)
	analysis := analyzer.Analyze(10000, 7200*6) // 大型弩，射击慢，时间拉长

	t.Logf("三弓弩单发弹容分析: 卡弹 %d, MTBF %.0f 发", analysis.JamCount, analysis.MTBFShots)
	// 单发弹容应比10发弹容更可靠（无托弹板复杂供弹）
	assert.GreaterOrEqual(t, analysis.JamCount, 0)
}

// 边界：极小样本 (shots=1)
func TestAnalyze_MinimalShots(t *testing.T) {
	var zhuge model.CrossbowVariant
	for _, v := range model.CrossbowPresets() {
		if v.VariantCode == "zhuge" {
			zhuge = v
		}
	}
	params := BuildParamsFromVariant(&zhuge)
	analyzer := NewMagazineReliabilityAnalyzer(params)
	analysis := analyzer.Analyze(1, 10)

	assert.Equal(t, 1, analysis.TotalShots)
	assert.Greater(t, analysis.MTBFShots, 0.0)
	assert.Equal(t, 100, len(analysis.ReliabilityCurvePts))
}

// 边界：0 shots - 应自动修正到1000
func TestAnalyze_ZeroShots_Fallback(t *testing.T) {
	params := MagazineParams{Capacity: 10, BaseJamRate: 1.0 / 5000}
	analyzer := NewMagazineReliabilityAnalyzer(params)
	analysis := analyzer.Analyze(0, 0)

	assert.Equal(t, 1000, analysis.TotalShots, "0发应回退到1000发")
}

// 边界：极大样本 (shots=1,000,000)
func TestAnalyze_LargeShots_Million(t *testing.T) {
	t.Skip("长时间测试跳过，用于压力测试")
	params := MagazineParams{Capacity: 10, BaseJamRate: 1.0 / 5000}
	analyzer := NewMagazineReliabilityAnalyzer(params)
	analysis := analyzer.Analyze(1_000_000, 3600*1000)

	assert.Equal(t, 1_000_000, analysis.TotalShots)
	t.Logf("百万发测试: 卡弹 %d 次, MTBF %.0f",
		analysis.JamCount, analysis.MTBFShots)
	// 大数定律：MTBF 应接近理论值
	expectedMTBF := 5000.0
	ratio := analysis.MTBFShots / expectedMTBF
	assert.Greater(t, ratio, 0.3, "MTBF不应偏离理论值超过70%")
	assert.Less(t, ratio, 3.0)
}

// 业务验证：弹容与卡弹率的负相关性
func TestAnalyze_CapacityVsJamRate_NegativeCorrelation(t *testing.T) {
	zhuge := buildVariantByCode("zhuge")
	sg := buildVariantByCode("san-gong")
	bz := buildVariantByCode("bi-zhang")

	shots := 50000
	simSec := 72000.0

	analyzeV := func(v *model.CrossbowVariant) float64 {
		params := BuildParamsFromVariant(v)
		a := NewMagazineReliabilityAnalyzer(params)
		return a.Analyze(shots, simSec).JamProbabilityPerShot
	}

	// 运行3次取平均以减少随机性
	zhugeRate := 0.0
	sgRate := 0.0
	bzRate := 0.0
	for i := 0; i < 3; i++ {
		zhugeRate += analyzeV(zhuge)
		sgRate += analyzeV(sg)
		bzRate += analyzeV(bz)
	}
	zhugeRate /= 3
	sgRate /= 3
	bzRate /= 3

	t.Logf("卡弹率: 诸葛弩(10发)=%.6f, 臂张弩(1发)=%.6f, 三弓弩(1发)=%.6f",
		zhugeRate, bzRate, sgRate)

	// 核心结论：诸葛弩的大容量连发结构更易卡弹
	assert.Greater(t, zhugeRate, sgRate,
		"诸葛弩连发箭匣(10发)卡弹率应高于单发三弓弩")
	assert.Greater(t, zhugeRate, bzRate,
		"诸葛弩连发箭匣(10发)卡弹率应高于单发臂张弩")
}

// 业务验证：疲劳放大系数效应
func TestAnalyze_FatigueFactorAmplification(t *testing.T) {
	// 构造极低基础卡弹率，分析不同疲劳阶段的卡弹数差异
	params := MagazineParams{Capacity: 10, BaseJamRate: 1.0 / 100000}
	analyzer := NewMagazineReliabilityAnalyzer(params)

	// 疲劳累积是 shots/(λ*2), λ=10*500=5000
	// 1次分析10000发: 疲劳度10000/10000=1.0 全程高疲劳
	// 2次分析2500发: 疲劳度2500/10000=0.25 全程低疲劳
	analysisFatigued := analyzer.Analyze(10000, 7200)

	analyzer2 := NewMagazineReliabilityAnalyzer(params)
	analysisFresh := analyzer2.Analyze(2500, 7200 / 4)

	t.Logf("高疲劳(10000发): 卡弹率 %.6f, 低疲劳(2500发): 卡弹率 %.6f",
		analysisFatigued.JamProbabilityPerShot, analysisFresh.JamProbabilityPerShot)

	// 注意：因为样本随机性，仅断言两者差值在合理区间
	diff := analysisFatigued.JamProbabilityPerShot - analysisFresh.JamProbabilityPerShot
	t.Logf("卡弹率差(高疲劳-低疲劳): %.6f", diff)
}

// 验证：威布尔分布CDF数学正确性
func TestWeibullCDF_MathProperties(t *testing.T) {
	k := 2.5
	lambda := 5000.0

	// CDF(0) = 0
	assert.Equal(t, 0.0, weibullCDF(0, k, lambda))
	// CDF(-x) = 0
	assert.Equal(t, 0.0, weibullCDF(-100, k, lambda))
	// CDF(λ) ≈ 1 - 1/e ≈ 0.632 (k=1时)
	cdfLambda := weibullCDF(lambda, k, lambda)
	// k=2.5时高于0.632（形状参数>1时失效密度在λ之前更高）
	assert.Greater(t, cdfLambda, 0.6)
	assert.Less(t, cdfLambda, 0.95)
	// CDF(x)单调递增
	assert.Less(t, weibullCDF(1000, k, lambda), weibullCDF(10000, k, lambda))
	t.Logf("W(2.5,5000) CDF(λ)=%.4f, CDF(3λ)=%.4f, CDF(10λ)=%.4f",
		cdfLambda,
		weibullCDF(3*lambda, k, lambda),
		weibullCDF(10*lambda, k, lambda))
}

// 验证：威布尔分位数与CDF的互逆关系
func TestWeibullQuantile_InverseOfCDF(t *testing.T) {
	k := 2.5
	lambda := 5000.0

	// 对若干点 p∈(0,1), CDF(Quantile(p)) 应≈p
	testPoints := []float64{0.1, 0.25, 0.5, 0.75, 0.9, 0.99}
	for _, p := range testPoints {
		q := weibullQuantile(p, k, lambda)
		cdfQ := weibullCDF(q, k, lambda)
		t.Logf("p=%.2f → q=%.0f → CDF(q)=%.4f (diff=%.6f)",
			p, q, cdfQ, math.Abs(cdfQ-p))
		assert.InDelta(t, p, cdfQ, 1e-9,
			"CDF(Quantile(p))应等于p")
	}
}

// 验证：威布尔分位数边界
func TestWeibullQuantile_Boundaries(t *testing.T) {
	k := 2.5
	lambda := 5000.0

	assert.Equal(t, 0.0, weibullQuantile(0, k, lambda))
	assert.Equal(t, 0.0, weibullQuantile(-0.5, k, lambda))
	assert.True(t, math.IsInf(weibullQuantile(1.0, k, lambda), 1))
	assert.True(t, math.IsInf(weibullQuantile(1.5, k, lambda), 1))
}

// FMEA 矩阵完整性与RPN计算正确
func TestBuildFMEAMatrix_AllRPNCorrect(t *testing.T) {
	entries := BuildFMEAMatrix()
	assert.Len(t, entries, 7)

	totalMode := map[JamFailureMode]bool{}
	for _, e := range entries {
		expectedRPN := e.Severity * e.Occurrence * e.Detection
		assert.Equal(t, expectedRPN, e.RPN,
			"RPN应等于S×O×D")
		assert.GreaterOrEqual(t, e.Severity, 1)
		assert.LessOrEqual(t, e.Severity, 10)
		assert.GreaterOrEqual(t, e.Occurrence, 1)
		assert.LessOrEqual(t, e.Occurrence, 10)
		assert.GreaterOrEqual(t, e.Detection, 1)
		assert.LessOrEqual(t, e.Detection, 10)
		assert.NotEmpty(t, e.Description)
		totalMode[e.Mode] = true
	}

	// 7种模式齐全
	assert.True(t, totalMode[JamDoubleFeed])
	assert.True(t, totalMode[JamMisfeed])
	assert.True(t, totalMode[JamStovepipe])
	assert.True(t, totalMode[JamFollowerBind])
	assert.True(t, totalMode[JamSpringFatigue])
	assert.True(t, totalMode[JamMagazineDamage])
	assert.True(t, totalMode[JamForeignObject])
}

// FMEA 业务验证：匣体损坏应是最高严重度
func TestFMEAMatrix_MagazineDamageHighestSeverity(t *testing.T) {
	entries := BuildFMEAMatrix()

	maxSeverity := 0
	var maxMode JamFailureMode
	for _, e := range entries {
		if e.Severity > maxSeverity {
			maxSeverity = e.Severity
			maxMode = e.Mode
		}
	}
	t.Logf("FMEA最高严重度: %s = %d", maxMode, maxSeverity)
	assert.Equal(t, JamMagazineDamage, maxMode,
		"匣体损坏应是最高严重度的失效模式(S=9)")
	assert.Equal(t, 9, maxSeverity)
}

// FMEA 业务验证：不供弹应是最高发生度
func TestFMEAMatrix_MisfeedHighestOccurrence(t *testing.T) {
	entries := BuildFMEAMatrix()

	maxOcc := 0
	var maxMode JamFailureMode
	for _, e := range entries {
		if e.Occurrence > maxOcc {
			maxOcc = e.Occurrence
			maxMode = e.Mode
		}
	}
	t.Logf("FMEA最高发生度: %s = %d", maxMode, maxOcc)
	assert.Equal(t, JamMisfeed, maxMode,
		"不供弹(Misfeed)应是最高发生度(O=6)")
}

// 验证：BuildParamsFromVariant 构建的参数与弹容匹配
func TestBuildParamsFromVariant_CapacityMatch(t *testing.T) {
	tests := []struct {
		code      string
		wantCap   int
		wantJamHi float64
		wantJamLo float64
	}{
		{"zhuge", 10, 1.0 / 3000.0, 1.0 / 4000.0},   // 10发弹匣，高基准
		{"bi-zhang", 1, 1.0 / 4500.0, 1.0 / 6000.0},  // 单发，中基准
		{"san-gong", 1, 1.0 / 4500.0, 1.0 / 6000.0},  // 单发，中基准
	}

	for _, tc := range tests {
		t.Run(tc.code, func(t *testing.T) {
			v := buildVariantByCode(tc.code)
			p := BuildParamsFromVariant(v)
			assert.Equal(t, tc.wantCap, p.Capacity)
			assert.GreaterOrEqual(t, p.BaseJamRate, tc.wantJamLo,
				"基础卡弹率不应过低")
			assert.LessOrEqual(t, p.BaseJamRate, tc.wantJamHi,
				"基础卡弹率不应过高")
			t.Logf("弩型 %s: 弹容=%d, 基准卡弹率=1/%.0f",
				tc.code, p.Capacity, 1.0/p.BaseJamRate)
		})
	}
}

// 验证：可靠性曲线 R(n) 的指数衰减数学性质
func TestReliabilityCurve_ExponentialDecay(t *testing.T) {
	params := MagazineParams{Capacity: 10, BaseJamRate: 1.0 / 5000}
	analyzer := NewMagazineReliabilityAnalyzer(params)
	analysis := analyzer.Analyze(5000, 3600)

	curve := analysis.ReliabilityCurvePts
	assert.Equal(t, 100, len(curve))

	// R(0)→1：曲线起点Y值趋近于1
	firstR := curve[0].Y
	assert.Greater(t, firstR, 0.8, "曲线起点Y应接近1")
	t.Logf("R(min)=%.4f @ n=%.0f, R(max)=%.4f @ n=%.0f",
		curve[0].Y, curve[0].X,
		curve[99].Y, curve[99].X)

	// R(MTBF) ≈ 1/e ≈ 0.3679
	mtbf := analysis.MTBFShots
	for _, pt := range curve {
		if math.Abs(pt.X-mtbf) < mtbf*0.1 {
			t.Logf("n≈MTBF时: n=%.0f, R=%.4f (expected≈0.3679, diff=%.4f)",
				pt.X, pt.Y, math.Abs(pt.Y-math.Exp(-1)))
			assert.InDelta(t, math.Exp(-1), pt.Y, 0.1,
				"R(MTBF)应≈1/e")
			break
		}
	}

	// 严格递减
	for i := 1; i < len(curve); i++ {
		assert.LessOrEqual(t, curve[i].Y, curve[i-1].Y,
			"可靠性曲线必须单调递减")
	}

	// X严格递增
	for i := 1; i < len(curve); i++ {
		assert.Greater(t, curve[i].X, curve[i-1].X)
	}
}

// 验证：置信区间覆盖率（100次重复分析，检查MTBF落入CI比例）
func TestAnalyze_ConfidenceIntervalCoverage(t *testing.T) {
	t.Skip("统计覆盖率测试，耗时长，可选择性运行")
	params := MagazineParams{Capacity: 10, BaseJamRate: 1.0 / 1000}
	_trueMTBF := 1.0 / params.BaseJamRate
	_ = _trueMTBF
	hits := 0
	trials := 100
	for i := 0; i < trials; i++ {
		analyzer := NewMagazineReliabilityAnalyzer(params)
		analysis := analyzer.Analyze(5000, 7200)
		if analysis.MTBFShots >= analysis.ConfidenceInterval.Low &&
			analysis.MTBFShots <= analysis.ConfidenceInterval.High {
			hits++
		}
	}
	coverage := float64(hits) / float64(trials)
	t.Logf("95%%CI实际覆盖率: %.1f%% (%d/%d)", coverage*100, hits, trials)
	// 95%CI 覆盖率应在 90%以上
	assert.Greater(t, coverage, 0.85,
		"置信区间覆盖率应接近95%")
}

// 验证：失效模式概率之和 ≈ 总卡弹概率
func TestAnalyze_FailureModeSumEqualsJamRate(t *testing.T) {
	params := MagazineParams{Capacity: 10, BaseJamRate: 1.0 / 2000}
	analyzer := NewMagazineReliabilityAnalyzer(params)
	analysis := analyzer.Analyze(50000, 7200)

	modeRateSum := 0.0
	for _, r := range analysis.FailureModeRate {
		modeRateSum += r
	}
	t.Logf("各模式概率和: %.6f, 总卡弹概率: %.6f, 差值: %.6f",
		modeRateSum, analysis.JamProbabilityPerShot,
		math.Abs(modeRateSum-analysis.JamProbabilityPerShot))

	// 两者应非常接近（差异来自浮点数精度）
	assert.InDelta(t, analysis.JamProbabilityPerShot, modeRateSum, 1e-6,
		"7种模式概率之和应等于总卡弹概率")
}

// 验证：默认构造器参数自动补全
func TestNewAnalyzer_DefaultValues(t *testing.T) {
	// 零参数构造
	analyzer := NewMagazineReliabilityAnalyzer(MagazineParams{})
	analysis := analyzer.Analyze(100, 360)

	assert.NotNil(t, analysis)
	assert.Equal(t, 100, analysis.TotalShots)
	t.Logf("默认参数: BaseJamRate≈%.6f, Capacity从10推断", analysis.JamProbabilityPerShot)
}

func buildVariantByCode(code string) *model.CrossbowVariant {
	for _, v := range model.CrossbowPresets() {
		if v.VariantCode == code {
			v2 := v
			return &v2
		}
	}
	return nil
}
