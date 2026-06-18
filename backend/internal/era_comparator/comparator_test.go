package era_comparator

import (
	"testing"
)

func TestNewComparator(t *testing.T) {
	c := NewComparator()
	if c == nil {
		t.Fatal("NewComparator() returned nil")
	}
	if len(c.ancient) != 3 {
		t.Errorf("expected 3 ancient, got %d", len(c.ancient))
	}
	if len(c.modernList) != 6 {
		t.Errorf("expected 6 modern firearms, got %d", len(c.modernList))
	}
}

func TestListAncient(t *testing.T) {
	c := NewComparator()
	list := c.ListAncient()
	if len(list) != 3 {
		t.Errorf("expected 3 ancient, got %d", len(list))
	}
}

func TestListModern(t *testing.T) {
	c := NewComparator()
	list := c.ListModern()
	if len(list) != 6 {
		t.Errorf("expected 6 modern, got %d", len(list))
	}
	for _, f := range list {
		if f.FirearmCode == "" {
			t.Error("modern firearm code should not be empty")
		}
		if f.Name == "" {
			t.Error("modern firearm name should not be empty")
		}
	}
}

func TestCompareNormalCase(t *testing.T) {
	c := NewComparator()
	res := c.Compare(CompareOptions{
		AncientCodes: []string{"zhuge", "san-gong"},
		ModernNames:  []string{"AK-47", "M249 SAW"},
		Metrics:      []string{"fireRate", "effectiveRange"},
	})
	if res == nil {
		t.Fatal("Compare returned nil")
	}
	if len(res.AncientVariants) != 2 {
		t.Errorf("expected 2 ancient variants, got %d", len(res.AncientVariants))
	}
	if len(res.ModernFirearms) != 2 {
		t.Errorf("expected 2 modern firearms, got %d", len(res.ModernFirearms))
	}
	if len(res.EraGapTable) != 5 {
		t.Errorf("expected 5 gap entries, got %d", len(res.EraGapTable))
	}
	for _, e := range res.EraGapTable {
		if e.GapRatio < 1 {
			t.Errorf("gap ratio for %s should be >=1, got %f", e.Metric, e.GapRatio)
		}
	}
}

func TestCompareDefaultOptions(t *testing.T) {
	c := NewComparator()
	res := c.Compare(CompareOptions{})
	if len(res.AncientVariants) != 3 {
		t.Errorf("default ancient should be 3, got %d", len(res.AncientVariants))
	}
	if len(res.ModernFirearms) != 2 {
		t.Errorf("default modern should be 2, got %d", len(res.ModernFirearms))
	}
	if len(res.EraGapTable) != 5 {
		t.Errorf("default gap table should be 5 rows, got %d", len(res.EraGapTable))
	}
}

func TestCompareInvalidCodes(t *testing.T) {
	c := NewComparator()
	res := c.Compare(CompareOptions{
		AncientCodes: []string{"fake1", "fake2"},
		ModernNames:  []string{"Nonexistent Gun", "X-999"},
	})
	if len(res.Errors) != 4 {
		t.Errorf("expected 4 error messages, got %d", len(res.Errors))
	}
	if res.BestAncient["fireRate"] == 0 {
		t.Error("should fallback to default ancient best values")
	}
	if res.BestModern["fireRate"] == 0 {
		t.Error("should fallback to default modern best values")
	}
}

func TestTechProgressAssertionFireRateGaps(t *testing.T) {
	c := NewComparator()
	rep := c.TechProgressAssertion()
	if rep == nil {
		t.Fatal("TechProgressAssertion returned nil")
	}
	if rep.Ak47ZhugeRatio < 5 {
		t.Errorf("AK47 vs zhuge fire rate ratio expected >= 5, got %f", rep.Ak47ZhugeRatio)
	}
	if rep.M249SanGongRatio < 50 {
		t.Errorf("M249 vs sangong ratio expected >= 50, got %f", rep.M249SanGongRatio)
	}
	if rep.M249TenMinRatio < 10 {
		t.Errorf("M249 10-min vs sangong ratio expected >= 10, got %f", rep.M249TenMinRatio)
	}
	if rep.GapOverall < 5 {
		t.Errorf("overall tech gap expected >= 5, got %f", rep.GapOverall)
	}
}

func TestTechProgressAssertionBoundaryValues(t *testing.T) {
	c := NewComparator()
	rep := c.TechProgressAssertion()
	if rep.AncientFireRateMax <= 0 {
		t.Error("ancient max fire rate should be positive")
	}
	if rep.ModernFireRateMax <= rep.AncientFireRateMax {
		t.Errorf("modern fire rate (%f) should exceed ancient (%f)", rep.ModernFireRateMax, rep.AncientFireRateMax)
	}
	if rep.AncientRangeMax <= 0 {
		t.Error("ancient max range should be positive")
	}
	if rep.ModernRangeMax <= rep.AncientRangeMax {
		t.Errorf("modern range (%f) should exceed ancient (%f)", rep.ModernRangeMax, rep.AncientRangeMax)
	}
}
