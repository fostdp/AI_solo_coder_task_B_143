package mechanism_comparator

import (
	"math"
	"sort"

	"crossbow-simulation/backend/internal/model"
)

type Comparator struct {
	variants map[string]model.CrossbowVariant
}

type CompareOptions struct {
	VariantCodes   []string
	CompareMetrics []string
}

type CompareResult struct {
	ComparedVariants []model.CrossbowVariant
	ComparedCodes    []string
	PerformanceRadar []model.PerformanceRadar
	AdvantageMap     []model.AdvantageMap
	Errors           []string
}

func NewComparator() *Comparator {
	vm := make(map[string]model.CrossbowVariant)
	for _, v := range model.CrossbowPresets() {
		vm[v.VariantCode] = v
	}
	return &Comparator{variants: vm}
}

func (c *Comparator) GetAll() []model.CrossbowVariant {
	return model.CrossbowPresets()
}

func (c *Comparator) GetByCode(code string) (model.CrossbowVariant, bool) {
	v, ok := c.variants[code]
	return v, ok
}

func metaValue(m *model.MeasurementMeta) float64 {
	if m == nil {
		return 0
	}
	return m.Value
}

func GetPerformanceMetric(v *model.CrossbowVariant, metric string) float64 {
	if v == nil {
		return 0
	}
	switch metric {
	case "drawWeight":
		return metaValue(v.Performance.DrawWeight)
	case "maxRange":
		return metaValue(v.Performance.MaxRange)
	case "effectiveRange":
		return metaValue(v.Performance.EffectiveRange)
	case "idealFireRate":
		return metaValue(v.Performance.IdealFireRate)
	case "magazineSize":
		return float64(v.Performance.MagazineSize)
	case "reloadTime":
		return metaValue(v.Performance.ReloadTime)
	case "accuracyScore":
		return metaValue(v.Performance.AccuracyScore)
	default:
		return 0
	}
}

func IsHigherBetter(metric string) bool {
	switch metric {
	case "reloadTime":
		return false
	default:
		return true
	}
}

func GetMetricLabel(metric string) string {
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

func DefaultMetrics() []string {
	return []string{"drawWeight", "maxRange", "effectiveRange", "idealFireRate", "magazineSize", "reloadTime", "accuracyScore"}
}

func DefaultVariantCodes() []string {
	return []string{"zhuge", "san-gong", "bi-zhang"}
}

type rankedKV struct {
	Code  string
	Value float64
}

func (c *Comparator) Compare(opts CompareOptions) (*CompareResult, error) {
	if len(opts.VariantCodes) == 0 {
		opts.VariantCodes = DefaultVariantCodes()
	}
	if len(opts.CompareMetrics) == 0 {
		opts.CompareMetrics = DefaultMetrics()
	}

	compared := make([]model.CrossbowVariant, 0, len(opts.VariantCodes))
	codeList := make([]string, 0)
	errMsgs := make([]string, 0)
	for _, code := range opts.VariantCodes {
		if v, ok := c.variants[code]; ok {
			compared = append(compared, v)
			codeList = append(codeList, code)
		} else {
			errMsgs = append(errMsgs, "无效弩型编码已忽略: "+code)
		}
	}
	if len(compared) == 0 {
		return &CompareResult{Errors: []string{"无有效弩型编码"}}, nil
	}

	radar := make([]model.PerformanceRadar, 0, len(opts.CompareMetrics))
	advantages := make([]model.AdvantageMap, 0, len(opts.CompareMetrics))

	for _, metric := range opts.CompareMetrics {
		vals := make(map[string]float64)
		higher := IsHigherBetter(metric)
		var ranked []rankedKV
		for i, v := range compared {
			val := GetPerformanceMetric(&v, metric)
			vals[codeList[i]] = val
			ranked = append(ranked, rankedKV{Code: codeList[i], Value: val})
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
			Metric: GetMetricLabel(metric),
			Values: vals,
			Best:   best.Code,
		})
		advantages = append(advantages, model.AdvantageMap{
			Metric:         GetMetricLabel(metric),
			BestVariant:    best.Code,
			BestValue:      best.Value,
			RunnerUp:       runnerUpCode,
			AdvantageRatio: ratio,
		})
	}

	return &CompareResult{
		ComparedVariants: compared,
		ComparedCodes:    codeList,
		PerformanceRadar: radar,
		AdvantageMap:     advantages,
		Errors:           errMsgs,
	}, nil
}

type FireRateReport struct {
	ZhugeRPM     float64
	BiZhangRPM   float64
	SanGongRPM   float64
	ZhugeBiRatio float64
	ZhugeSanRatio float64
	BiSanRatio   float64
	RankKing     string
}

func (c *Comparator) FireRateAssertion() *FireRateReport {
	zhuge, _ := c.GetByCode("zhuge")
	bizhang, _ := c.GetByCode("bi-zhang")
	sangong, _ := c.GetByCode("san-gong")
	zR := GetPerformanceMetric(&zhuge, "idealFireRate")
	bR := GetPerformanceMetric(&bizhang, "idealFireRate")
	sR := GetPerformanceMetric(&sangong, "idealFireRate")
	king := "zhuge"
	if !(zR >= bR && bR >= sR) {
		king = "unexpected"
	}
	return &FireRateReport{
		ZhugeRPM:      zR,
		BiZhangRPM:    bR,
		SanGongRPM:    sR,
		ZhugeBiRatio:  zR / math.Max(0.0001, bR),
		ZhugeSanRatio: zR / math.Max(0.0001, sR),
		BiSanRatio:    bR / math.Max(0.0001, sR),
		RankKing:      king,
	}
}
