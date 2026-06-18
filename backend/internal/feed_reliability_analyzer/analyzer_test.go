package feed_reliability_analyzer

import (
	"math"
	"testing"

	"crossbow-simulation/backend/internal/model"
)

func TestWeibullCDF(t *testing.T) {
	k := 2.5
	lambda := 5000.0
	if WeibullCDF(0, k, lambda) != 0 {
		t.Error("CDF at 0 should be 0")
	}
	if WeibullCDF(-5, k, lambda) != 0 {
		t.Error("CDF at negative x should be 0")
	}
	p1000 := WeibullCDF(1000, k, lambda)
	if p1000 < 0 || p1000 > 1 {
		t.Errorf("CDF at 1000 should be in [0,1], got %f", p1000)
	}
	p10000 := WeibullCDF(10000, k, lambda)
	if p10000 < 0.95 {
		t.Errorf("CDF at 10000 should be near 1, got %f", p10000)
	}
	if p10000 < p1000 {
		t.Error("CDF should be monotonically increasing")
	}
}

func TestWeibullCDFMonotonicity(t *testing.T) {
	k := 2.5
	lambda := 10000.0
	prev := -1.0
	for x := 0; x <= 50000; x += 1000 {
		v := WeibullCDF(float64(x), k, lambda)
		if v < prev {
			t.Errorf("CDF not monotonic at x=%d: %f < %f", x, v, prev)
		}
		prev = v
	}
}

func TestWeibullQuantileInverse(t *testing.T) {
	k := 2.5
	lambda := 10000.0
	for p := 0.05; p < 0.95; p += 0.05 {
		x := WeibullQuantile(p, k, lambda)
		cdf := WeibullCDF(x, k, lambda)
		if math.Abs(cdf-p) > 1e-9 {
			t.Errorf("Quantile inverse failed at p=%f: x=%f -> cdf=%f, diff=%.2e", p, x, cdf, cdf-p)
		}
	}
}

func TestWeibullQuantileBoundary(t *testing.T) {
	k := 2.5
	lambda := 5000.0
	if WeibullQuantile(0, k, lambda) != 0 {
		t.Error("Quantile at p=0 should be 0")
	}
	if !math.IsInf(WeibullQuantile(1, k, lambda), 1) {
		t.Error("Quantile at p=1 should be +Inf")
	}
}

func TestGetFrictionCoefficientDatabase(t *testing.T) {
	db := GetFrictionCoefficientDatabase()
	if len(db) != 8 {
		t.Errorf("expected 8 friction measurements, got %d", len(db))
	}
	for i, m := range db {
		if m.MaterialPair == "" {
			t.Errorf("measurement %d: material pair empty", i)
		}
		if m.MeanCoeff <= 0 {
			t.Errorf("measurement %d: mean coeff should be positive, got %f", i, m.MeanCoeff)
		}
		if m.MeanCoeff < m.Low95CI || m.MeanCoeff > m.High95CI {
			t.Errorf("measurement %d: mean %f not in [%f, %f]", i, m.MeanCoeff, m.Low95CI, m.High95CI)
		}
		if m.SampleCount <= 0 {
			t.Errorf("measurement %d: sample count should be positive", i)
		}
		if m.Source == "" {
			t.Errorf("measurement %d: source empty", i)
		}
	}
}

func TestLookupFriction(t *testing.T) {
	mu, ref := LookupFriction("竹", FrictionDryClean)
	if mu <= 0 {
		t.Error("bamboo lookup should return positive mu")
	}
	if ref == nil {
		t.Error("bamboo lookup should return ref measurement")
	}

	mu2, ref2 := LookupFriction("铁木", FrictionDryClean)
	if mu2 <= mu {
		t.Errorf("ironwood should have higher mu than bamboo: %f vs %f", mu2, mu)
	}
	if ref2 == nil {
		t.Error("ironwood lookup should return ref")
	}

	_, ref3 := LookupFriction("", FrictionDryClean)
	if ref3 == nil {
		t.Error("empty substring should return default")
	}

	mu4, _ := LookupFriction("NONEXISTENTMATERIAL", FrictionDryClean)
	if mu4 <= 0 {
		t.Error("unknown material should fallback to default positive value")
	}
}

func TestBuildFMEAMatrix(t *testing.T) {
	fmea := BuildFMEAMatrix()
	if len(fmea) != 7 {
		t.Errorf("expected 7 FMEA entries, got %d", len(fmea))
	}
	maxRPN := 0
	for _, e := range fmea {
		if e.Severity < 1 || e.Severity > 10 {
			t.Errorf("severity out of 1-10 range: %d for %s", e.Severity, e.Mode)
		}
		if e.Occurrence < 1 || e.Occurrence > 10 {
			t.Errorf("occurrence out of 1-10: %d for %s", e.Occurrence, e.Mode)
		}
		if e.Detection < 1 || e.Detection > 10 {
			t.Errorf("detection out of 1-10: %d for %s", e.Detection, e.Mode)
		}
		expectedRPN := e.Severity * e.Occurrence * e.Detection
		if e.RPN != expectedRPN {
			t.Errorf("RPN mismatch for %s: got %d, expected %d", e.Mode, e.RPN, expectedRPN)
		}
		if e.RPN > maxRPN {
			maxRPN = e.RPN
		}
	}
	if maxRPN <= 0 {
		t.Error("max RPN should be positive")
	}
}

func TestBuildParamsFromVariant(t *testing.T) {
	zhuge := createTestVariant("zhuge", 10, 0.14)
	p := BuildParamsFromVariant(zhuge)
	if p.Capacity != 10 {
		t.Errorf("zhuge capacity should be 10, got %d", p.Capacity)
	}
	if p.FollowerFriction <= 0 {
		t.Error("zhuge friction should be positive")
	}
	if p.BaseJamRate <= 0 {
		t.Error("zhuge jam rate should be positive")
	}

	sangong := createTestVariant("san-gong", 1, 0.18)
	p2 := BuildParamsFromVariant(sangong)
	if p2.Capacity != 1 {
		t.Errorf("sangong capacity should be 1, got %d", p2.Capacity)
	}
	if p2.BaseJamRate > p.BaseJamRate {
		t.Errorf("sangong (single) should have lower jam rate than zhuge (10): %f vs %f", p2.BaseJamRate, p.BaseJamRate)
	}

	p3 := BuildParamsFromVariant(nil)
	if p3.Capacity <= 0 {
		t.Error("nil variant should still give valid capacity")
	}
}

func TestAnalyzeDeterministic(t *testing.T) {
	params := MagazineParams{
		Capacity:         10,
		SpringRate:       850,
		FollowerFriction: 0.14,
		BaseJamRate:      1.0 / 3500.0,
		ToleranceClass:   "standard",
		EnvironmentalCond: FrictionDryClean,
	}
	analyzer := NewAnalyzer(params)
	if analyzer == nil {
		t.Fatal("NewAnalyzer returned nil")
	}
	res := analyzer.Analyze(10000, 7200)
	if res == nil {
		t.Fatal("Analyze returned nil")
	}
	if res.TotalShots != 10000 {
		t.Errorf("total shots mismatch: %d vs 10000", res.TotalShots)
	}
	if res.JamProbabilityPerShot < 0 || res.JamProbabilityPerShot > 1 {
		t.Errorf("jam prob per shot out of [0,1]: %f", res.JamProbabilityPerShot)
	}
	if res.MTBFShots < 100 {
		t.Errorf("MTBF should be reasonable, got %f", res.MTBFShots)
	}
	if len(res.ReliabilityCurvePts) < 10 {
		t.Errorf("reliability curve should have >= 10 points, got %d", len(res.ReliabilityCurvePts))
	}
	if len(res.FMEAMatrix) != 7 {
		t.Errorf("FMEA matrix should have 7 entries, got %d", len(res.FMEAMatrix))
	}
	if len(res.FailureModeDistribution) < 7 {
		t.Errorf("failure mode dist should cover 7 modes, got %d", len(res.FailureModeDistribution))
	}
	if res.FrictionCoeffUsed != params.FollowerFriction {
		t.Errorf("friction coeff mismatch: %f vs %f", res.FrictionCoeffUsed, params.FollowerFriction)
	}
}

func TestAnalyzeWeibullShapeAndScale(t *testing.T) {
	params := MagazineParams{Capacity: 10, BaseJamRate: 1.0 / 3500.0}
	analyzer := NewAnalyzer(params)
	res := analyzer.Analyze(50000, 36000)
	if res.WeibullShapeK < 2.0 || res.WeibullShapeK > 3.0 {
		t.Errorf("Weibull k should be ~2.5, got %f", res.WeibullShapeK)
	}
	// Capacity=10 → λ = capacity*500 = 5000
	if res.WeibullScaleLambda < 3000 {
		t.Errorf("Weibull lambda should be >= 3000 (capacity*500), got %f", res.WeibullScaleLambda)
	}
	if res.WeibullScaleLambda > 100000 {
		t.Errorf("Weibull lambda should be < 100000, got %f", res.WeibullScaleLambda)
	}
}

func TestAnalyzeReliabilityCurveMonotone(t *testing.T) {
	analyzer := NewAnalyzer(MagazineParams{Capacity: 10})
	res := analyzer.Analyze(100000, 72000)
	if len(res.ReliabilityCurvePts) < 2 {
		t.Fatal("not enough curve points")
	}
	prev := 1.0
	for i, pt := range res.ReliabilityCurvePts {
		if pt.Y > prev+1e-9 {
			t.Errorf("curve %d: y=%f > prev=%f (should decay)", i, pt.Y, prev)
		}
		if pt.Y < 0 || pt.Y > 1 {
			t.Errorf("curve %d y out of [0,1]: %f", i, pt.Y)
		}
		prev = pt.Y
	}
	if prev > 0.99 {
		t.Error("by end of curve reliability should have decayed somewhat")
	}
}

func TestGenerateReport(t *testing.T) {
	zhuge := createTestVariant("zhuge", 10, 0.14)
	params := BuildParamsFromVariant(zhuge)
	analyzer := NewAnalyzer(params)
	analysis := analyzer.Analyze(50000, 36000)
	report := analyzer.GenerateReport("zhuge", analysis)
	if report == nil {
		t.Fatal("GenerateReport returned nil")
	}
	if report.VariantCode != "zhuge" {
		t.Errorf("variant code mismatch: %s", report.VariantCode)
	}
	if report.CapacityJamRating != "high" {
		t.Errorf("zhuge (10-capacity) should rate 'high', got %s", report.CapacityJamRating)
	}
	if report.MuJamAmplification < 1.0 {
		t.Errorf("mu amplification should be >= 1, got %f", report.MuJamAmplification)
	}
	if report.ReliabilityAtMTBF < 0.2 || report.ReliabilityAtMTBF > 0.5 {
		t.Errorf("R(MTBF) should be ~0.3679 (1/e), got %f (range [0.2,0.5])", report.ReliabilityAtMTBF)
	}
	if !report.CurveStrictDecay {
		t.Error("curve should be strictly decaying")
	}
	if report.RPNWorst <= 0 {
		t.Error("worst RPN should be positive")
	}
}

func TestAnalyzeConcurrent(t *testing.T) {
	analyzer := NewAnalyzer(MagazineParams{Capacity: 10, BaseJamRate: 1.0 / 3500.0})
	done := make(chan struct{})
	for i := 0; i < 4; i++ {
		go func(seed int) {
			n := 2000 + seed*1000
			r := analyzer.Analyze(n, 3600)
			if r == nil || r.TotalShots != n {
				t.Errorf("concurrent analyze expected %d shots", n)
			}
			done <- struct{}{}
		}(i)
	}
	for i := 0; i < 4; i++ {
		<-done
	}
}

func createTestVariant(code string, capacity int, frictionVal float64) *model.CrossbowVariant {
	return &model.CrossbowVariant{
		VariantCode: code,
		Performance: model.PerformanceMetrics{MagazineSize: capacity},
		MechanismParams: &model.MechanismParams{
			FrictionCoeff: &model.MeasurementMeta{Value: frictionVal},
		},
	}
}
