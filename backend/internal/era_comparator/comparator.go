package era_comparator

import (
	"math"

	"crossbow-simulation/backend/internal/model"
	"crossbow-simulation/backend/internal/mechanism_comparator"
)

type Comparator struct {
	ancient  map[string]model.CrossbowVariant
	modern   map[string]model.ModernFirearm
	modernList []model.ModernFirearm
}

type CompareOptions struct {
	AncientCodes []string
	ModernNames  []string
	Metrics      []string
}

type CompareResult struct {
	AncientVariants []model.CrossbowVariant
	ModernFirearms  []model.ModernFirearm
	EraGapTable     []model.EraGapEntry
	BestAncient     map[string]float64
	BestModern      map[string]float64
	Errors          []string
}

type TechProgressReport struct {
	Ak47ZhugeRatio     float64
	M249SanGongRatio   float64
	M249TenMinRatio    float64
	AncientFireRateMax float64
	AncientRangeMax    float64
	ModernFireRateMax  float64
	ModernRangeMax     float64
	GapOverall         float64
}

func NewComparator() *Comparator {
	a := make(map[string]model.CrossbowVariant)
	for _, v := range model.CrossbowPresets() {
		a[v.VariantCode] = v
	}
	m := make(map[string]model.ModernFirearm)
	list := model.ModernFirearmPresets()
	for _, f := range list {
		m[f.FirearmCode] = f
		m[f.Name] = f
	}
	return &Comparator{ancient: a, modern: m, modernList: list}
}

func (c *Comparator) ListAncient() []model.CrossbowVariant {
	return model.CrossbowPresets()
}

func (c *Comparator) ListModern() []model.ModernFirearm {
	return c.modernList
}

func DefaultAncientCodes() []string {
	return []string{"zhuge", "san-gong", "bi-zhang"}
}

func DefaultModernNames() []string {
	return []string{"AK-47", "M249 SAW"}
}

func DefaultMetrics() []string {
	return []string{"fireRate", "effectiveRange", "magazineSize", "muzzleVelocity"}
}

func (c *Comparator) Compare(opts CompareOptions) *CompareResult {
	if len(opts.AncientCodes) == 0 {
		opts.AncientCodes = DefaultAncientCodes()
	}
	if len(opts.ModernNames) == 0 {
		opts.ModernNames = DefaultModernNames()
	}
	if len(opts.Metrics) == 0 {
		opts.Metrics = DefaultMetrics()
	}

	errs := make([]string, 0)
	ancientList := make([]model.CrossbowVariant, 0)
	for _, code := range opts.AncientCodes {
		if v, ok := c.ancient[code]; ok {
			ancientList = append(ancientList, v)
		} else {
			errs = append(errs, "忽略无效古代弩型: "+code)
		}
	}
	modernList := make([]model.ModernFirearm, 0)
	for _, n := range opts.ModernNames {
		if f, ok := c.modern[n]; ok {
			modernList = append(modernList, f)
		} else {
			errs = append(errs, "忽略无效现代武器: "+n)
		}
	}

	bestAncient := map[string]float64{"fireRate": 0, "effectiveRange": 0, "magazineSize": 0}
	for _, v := range ancientList {
		fr := mechanism_comparator.GetPerformanceMetric(&v, "idealFireRate")
		er := mechanism_comparator.GetPerformanceMetric(&v, "effectiveRange")
		mg := float64(v.Performance.MagazineSize)
		if fr > bestAncient["fireRate"] { bestAncient["fireRate"] = fr }
		if er > bestAncient["effectiveRange"] { bestAncient["effectiveRange"] = er }
		if mg > bestAncient["magazineSize"] { bestAncient["magazineSize"] = mg }
	}

	bestModern := map[string]float64{"fireRate": 0, "effectiveRange": 0, "magazineSize": 0, "muzzleVelocity": 0}
	for _, f := range modernList {
		fr := f.EffectiveRPM
		if fr == 0 {
			fr = f.CyclicRateRPM * 0.15
		}
		if fr > bestModern["fireRate"] { bestModern["fireRate"] = fr }
		if f.EffectiveRangeM > bestModern["effectiveRange"] { bestModern["effectiveRange"] = f.EffectiveRangeM }
		if float64(f.MagazineSize) > bestModern["magazineSize"] { bestModern["magazineSize"] = float64(f.MagazineSize) }
		if f.MuzzleVelocityMPS > bestModern["muzzleVelocity"] { bestModern["muzzleVelocity"] = f.MuzzleVelocityMPS }
	}

	// fallback default
	if bestAncient["fireRate"] <= 0 { bestAncient["fireRate"] = 10 }
	if bestAncient["effectiveRange"] <= 0 { bestAncient["effectiveRange"] = 300 }
	if bestAncient["magazineSize"] == 0 { bestAncient["magazineSize"] = 10 }
	if bestModern["fireRate"] <= 0 { bestModern["fireRate"] = 100 }
	if bestModern["effectiveRange"] <= 0 { bestModern["effectiveRange"] = 400 }
	if bestModern["magazineSize"] == 0 { bestModern["magazineSize"] = 30 }
	if bestModern["muzzleVelocity"] <= 0 { bestModern["muzzleVelocity"] = 700 }

	ancientMV := 80.0   // 古箭平均初速 m/s
	ancientKEst := 160.0 // 古箭平均动能 J
	modernKEst := 2000.0 // 现代弹平均动能 J

	gapTable := []model.EraGapEntry{
		{
			Metric:       "射速(实战)",
			AncientValue: bestAncient["fireRate"], AncientUnit: "发/分",
			ModernValue: bestModern["fireRate"], ModernUnit: "发/分",
			GapRatio:    bestModern["fireRate"] / math.Max(1e-6, bestAncient["fireRate"]),
			Remark:      "自动武器循环射速使火力密度提升数量级",
		},
		{
			Metric:       "有效射程",
			AncientValue: bestAncient["effectiveRange"], AncientUnit: "米",
			ModernValue: bestModern["effectiveRange"], ModernUnit: "米",
			GapRatio:    bestModern["effectiveRange"] / math.Max(1, bestAncient["effectiveRange"]),
			Remark:      "线膛枪管与定装弹显著延伸有效射程",
		},
		{
			Metric:       "弹容",
			AncientValue: bestAncient["magazineSize"], AncientUnit: "发",
			ModernValue: bestModern["magazineSize"], ModernUnit: "发",
			GapRatio:    bestModern["magazineSize"] / math.Max(1, bestAncient["magazineSize"]),
			Remark:      "盒式弹匣与弹链使持续火力大幅提高",
		},
		{
			Metric:       "初速",
			AncientValue: ancientMV, AncientUnit: "m/s",
			ModernValue: bestModern["muzzleVelocity"], ModernUnit: "m/s",
			GapRatio:    bestModern["muzzleVelocity"] / math.Max(1, ancientMV),
			Remark:      "火药燃气提供的能量远超人力张弩",
		},
		{
			Metric:       "杀伤动能(估算)",
			AncientValue: ancientKEst, AncientUnit: "J",
			ModernValue: modernKEst, ModernUnit: "J",
			GapRatio:    modernKEst / math.Max(1, ancientKEst),
			Remark:      "基于箭重50g@80m/s vs 7.62mm弹头8g@715m/s估算",
		},
	}

	return &CompareResult{
		AncientVariants: ancientList,
		ModernFirearms:  modernList,
		EraGapTable:     gapTable,
		BestAncient:     bestAncient,
		BestModern:      bestModern,
		Errors:          errs,
	}
}

func (c *Comparator) TechProgressAssertion() *TechProgressReport {
	zhuge, _ := c.ancient["zhuge"]
	sangong, _ := c.ancient["san-gong"]
	ak, _ := c.modern["AK-47"]
	m249, _ := c.modern["M249 SAW"]

	zFR := mechanism_comparator.GetPerformanceMetric(&zhuge, "idealFireRate")
	sFR := mechanism_comparator.GetPerformanceMetric(&sangong, "idealFireRate")

	akFR := ak.EffectiveRPM
	if akFR == 0 { akFR = ak.CyclicRateRPM * 0.15 }
	m249FR := m249.EffectiveRPM
	if m249FR == 0 { m249FR = m249.CyclicRateRPM * 0.2 }

	// 10分钟战斗总发数（考虑装弹/换弹匣）
	m249TenMin := m249FR * 10.0
	sanTenMin := sFR * 10.0

	aFRMax := 0.0
	aRMax := 0.0
	for _, v := range c.ancient {
		if mechanism_comparator.GetPerformanceMetric(&v, "idealFireRate") > aFRMax {
			aFRMax = mechanism_comparator.GetPerformanceMetric(&v, "idealFireRate")
		}
		if mechanism_comparator.GetPerformanceMetric(&v, "effectiveRange") > aRMax {
			aRMax = mechanism_comparator.GetPerformanceMetric(&v, "effectiveRange")
		}
	}
	mFRMax := 0.0
	mRMax := 0.0
	for _, f := range c.modernList {
		fr := f.EffectiveRPM
		if fr == 0 { fr = f.CyclicRateRPM * 0.15 }
		if fr > mFRMax { mFRMax = fr }
		if f.EffectiveRangeM > mRMax { mRMax = f.EffectiveRangeM }
	}

	gap := (mFRMax/math.Max(1,aFRMax) + mRMax/math.Max(1,aRMax) + bestModernMag(c.modernList)/bestAncientMag(c.ancient)) / 3.0

	return &TechProgressReport{
		Ak47ZhugeRatio:     akFR / math.Max(1e-6, zFR),
		M249SanGongRatio:   m249FR / math.Max(1e-6, sFR),
		M249TenMinRatio:    m249TenMin / math.Max(1, sanTenMin),
		AncientFireRateMax: aFRMax,
		AncientRangeMax:    aRMax,
		ModernFireRateMax:  mFRMax,
		ModernRangeMax:     mRMax,
		GapOverall:         gap,
	}
}

func bestModernMag(list []model.ModernFirearm) float64 {
	best := 0.0
	for _, f := range list {
		if float64(f.MagazineSize) > best { best = float64(f.MagazineSize) }
	}
	if best <= 0 { best = 30 }
	return best
}

func bestAncientMag(mp map[string]model.CrossbowVariant) float64 {
	best := 0.0
	for _, v := range mp {
		if float64(v.Performance.MagazineSize) > best { best = float64(v.Performance.MagazineSize) }
	}
	if best <= 0 { best = 10 }
	return best
}
