package mechanism_comparator

import (
	"testing"
)

func TestNewComparator(t *testing.T) {
	c := NewComparator()
	if c == nil {
		t.Fatal("NewComparator() returned nil")
	}
	if len(c.variants) != 3 {
		t.Errorf("expected 3 variants, got %d", len(c.variants))
	}
}

func TestGetAll(t *testing.T) {
	c := NewComparator()
	all := c.GetAll()
	if len(all) != 3 {
		t.Errorf("GetAll() expected 3 items, got %d", len(all))
	}
	codes := map[string]bool{}
	for _, v := range all {
		codes[v.VariantCode] = true
	}
	for _, expected := range []string{"zhuge", "san-gong", "bi-zhang"} {
		if !codes[expected] {
			t.Errorf("GetAll() missing expected code: %s", expected)
		}
	}
}

func TestGetByCode(t *testing.T) {
	c := NewComparator()
	v, ok := c.GetByCode("zhuge")
	if !ok {
		t.Fatal("GetByCode(zhuge) should exist")
	}
	if v.VariantCode != "zhuge" {
		t.Errorf("unexpected code: %s", v.VariantCode)
	}
	if _, ok := c.GetByCode("nonexistent"); ok {
		t.Error("GetByCode(nonexistent) should return false")
	}
}

func TestGetPerformanceMetric(t *testing.T) {
	c := NewComparator()
	zhuge, _ := c.GetByCode("zhuge")
	fr := GetPerformanceMetric(&zhuge, "idealFireRate")
	if fr <= 0 {
		t.Errorf("zhuge idealFireRate should be > 0, got %f", fr)
	}
	if GetPerformanceMetric(&zhuge, "unknown_metric") != 0 {
		t.Error("unknown metric should return 0")
	}
	if GetPerformanceMetric(nil, "idealFireRate") != 0 {
		t.Error("nil variant should return 0")
	}
}

func TestIsHigherBetter(t *testing.T) {
	if !IsHigherBetter("idealFireRate") {
		t.Error("idealFireRate should be higher-is-better")
	}
	if IsHigherBetter("reloadTime") {
		t.Error("reloadTime should be lower-is-better")
	}
	if !IsHigherBetter("default") {
		t.Error("default should be higher-is-better")
	}
}

func TestCompareNormalCase(t *testing.T) {
	c := NewComparator()
	opts := CompareOptions{
		VariantCodes:   []string{"zhuge", "bi-zhang", "san-gong"},
		CompareMetrics: []string{"idealFireRate", "effectiveRange", "reloadTime"},
	}
	res, err := c.Compare(opts)
	if err != nil {
		t.Fatalf("Compare returned error: %v", err)
	}
	if res == nil {
		t.Fatal("Compare returned nil result")
	}
	if len(res.ComparedVariants) != 3 {
		t.Errorf("expected 3 compared variants, got %d", len(res.ComparedVariants))
	}
	if len(res.PerformanceRadar) != 3 {
		t.Errorf("expected 3 radar entries, got %d", len(res.PerformanceRadar))
	}
	if len(res.AdvantageMap) != 3 {
		t.Errorf("expected 3 advantage entries, got %d", len(res.AdvantageMap))
	}
	for _, adv := range res.AdvantageMap {
		if adv.BestVariant == "" {
			t.Error("advantage BestVariant should not be empty")
		}
		if adv.AdvantageRatio < 1 {
			t.Errorf("advantage ratio should be >= 1, got %f for %s", adv.AdvantageRatio, adv.Metric)
		}
	}
}

func TestCompareInvalidCodeIgnored(t *testing.T) {
	c := NewComparator()
	opts := CompareOptions{
		VariantCodes:   []string{"zhuge", "invalid_code"},
		CompareMetrics: []string{"idealFireRate"},
	}
	res, _ := c.Compare(opts)
	if len(res.ComparedVariants) != 1 {
		t.Errorf("should have only 1 valid variant, got %d", len(res.ComparedVariants))
	}
	if len(res.Errors) == 0 {
		t.Error("should report error for invalid code")
	}
}

func TestCompareEmptyCodes(t *testing.T) {
	c := NewComparator()
	res, _ := c.Compare(CompareOptions{})
	if len(res.ComparedVariants) != 3 {
		t.Errorf("empty options should use default 3 variants, got %d", len(res.ComparedVariants))
	}
}

func TestCompareAllInvalid(t *testing.T) {
	c := NewComparator()
	res, _ := c.Compare(CompareOptions{VariantCodes: []string{"a", "b", "c"}})
	if len(res.ComparedVariants) != 0 {
		t.Errorf("all-invalid should compare 0 variants, got %d", len(res.ComparedVariants))
	}
	if len(res.Errors) != 1 || res.Errors[0] != "无有效弩型编码" {
		t.Error("should report no valid variant error")
	}
}

func TestFireRateAssertion(t *testing.T) {
	c := NewComparator()
	report := c.FireRateAssertion()
	if report == nil {
		t.Fatal("FireRateAssertion returned nil")
	}
	if report.RankKing != "zhuge" {
		t.Errorf("zhuge should have highest fire rate rank, got %s", report.RankKing)
	}
	if report.ZhugeRPM <= report.BiZhangRPM {
		t.Errorf("zhuge RPM (%f) should be > bizhang RPM (%f)", report.ZhugeRPM, report.BiZhangRPM)
	}
	if report.BiZhangRPM <= report.SanGongRPM {
		t.Errorf("bizhang RPM (%f) should be > sangong RPM (%f)", report.BiZhangRPM, report.SanGongRPM)
	}
	if report.ZhugeSanRatio < 3 {
		t.Errorf("zhuge vs sangong ratio should be >= 3, got %f", report.ZhugeSanRatio)
	}
	if report.ZhugeBiRatio < 1.5 {
		t.Errorf("zhuge vs bizhang ratio should be >= 1.5, got %f", report.ZhugeBiRatio)
	}
}
