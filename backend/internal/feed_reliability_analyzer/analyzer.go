package feed_reliability_analyzer

import (
	"math"
	"math/rand"
	"sync"
	"time"

	"crossbow-simulation/backend/internal/model"
)

type JamFailureMode string

const (
	JamDoubleFeed     JamFailureMode = "DoubleFeed"
	JamMisfeed        JamFailureMode = "Misfeed"
	JamStovepipe      JamFailureMode = "Stovepipe"
	JamFollowerBind   JamFailureMode = "FollowerBind"
	JamSpringFatigue  JamFailureMode = "SpringFatigue"
	JamMagazineDamage JamFailureMode = "MagazineDamage"
	JamForeignObject  JamFailureMode = "ForeignObject"
)

type JamEvent struct {
	Mode            JamFailureMode `json:"mode"`
	Timestamp       int64          `json:"timestamp"`
	Severity        float64        `json:"severity"`
	Recoverable     bool           `json:"recoverable"`
	RootCause       string         `json:"rootCause"`
	RecoveryTimeSec float64        `json:"recoveryTimeSec"`
}

type ConfidenceInterval struct {
	Low   float64 `json:"low"`
	High  float64 `json:"high"`
	Level float64 `json:"level"`
}

type FMEAEntry struct {
	Mode        JamFailureMode `json:"mode"`
	Severity    int            `json:"severity"`
	Occurrence  int            `json:"occurrence"`
	Detection   int            `json:"detection"`
	RPN         int            `json:"rpn"`
	Description string         `json:"description"`
}

type FrictionCondition string

const (
	FrictionDryClean       FrictionCondition = "dry_clean"
	FrictionDryDusty       FrictionCondition = "dry_dusty"
	FrictionLubricated     FrictionCondition = "lubricated"
	FrictionHumid90RH      FrictionCondition = "humid_90rh"
	FrictionLowTempMinus20 FrictionCondition = "low_temp_-20c"
)

type FrictionMeasurement struct {
	MaterialPair    string            `json:"materialPair"`
	Condition       FrictionCondition `json:"condition"`
	MeanCoeff       float64           `json:"meanCoeff"`
	StdDev          float64           `json:"stdDev"`
	Low95CI         float64           `json:"low95CI"`
	High95CI        float64           `json:"high95CI"`
	Source          string            `json:"source"`
	MeasurementYear int               `json:"measurementYear"`
	SampleCount     int               `json:"sampleCount"`
	Method          string            `json:"method"`
	Notes           string            `json:"notes,omitempty"`
}

type MagazineParams struct {
	Capacity          int
	SpringRate        float64
	FollowerFriction  float64
	FrictionSource    *FrictionMeasurement
	ToleranceClass    string
	BaseJamRate       float64
	EnvironmentalCond FrictionCondition
}

type MagazineReliabilityAnalysis struct {
	TotalShots              int                `json:"totalShots"`
	JamCount                int                `json:"jamCount"`
	JamProbabilityPerShot   float64            `json:"jamProbabilityPerShot"`
	MTBFShots               float64            `json:"mtbfShots"`
	MTBFHours               float64            `json:"mtbfHours"`
	FailureModeDistribution map[string]int     `json:"failureModeDistribution"`
	FailureModeRate         map[string]float64 `json:"failureModeRate"`
	ReliabilityCurvePts     []model.Point      `json:"reliabilityCurvePts"`
	ConfidenceInterval      ConfidenceInterval `json:"confidenceInterval"`
	JamEvents               []JamEvent         `json:"jamEvents,omitempty"`
	FMEAMatrix              []FMEAEntry        `json:"fmeaMatrix"`
	FrictionCoeffUsed       float64            `json:"frictionCoeffUsed"`
	FrictionMeasurementRef  *FrictionMeasurement `json:"frictionMeasurementRef,omitempty"`
	EnvironmentalCondition  FrictionCondition  `json:"environmentalCondition"`
	WeibullShapeK           float64            `json:"weibullShapeK"`
	WeibullScaleLambda      float64            `json:"weibullScaleLambda"`
}

type ReliabilityReport struct {
	VariantCode       string
	CapacityJamRating string  // "low"/"medium"/"high" 按弹容区分
	MuJamAmplification float64 // 摩擦引起的卡弹率放大倍数
	R95LowRatio       float64 // CI 下限 / MTBF
	R95HighRatio      float64
	ModeMaxOccurrence JamFailureMode
	SeverityMaxMode   JamFailureMode
	RPNWorst          int
	ReliabilityAtMTBF float64
	CurveStrictDecay  bool
}

// ---- 摩擦系数数据库 ----

func GetFrictionCoefficientDatabase() []FrictionMeasurement {
	return []FrictionMeasurement{
		{MaterialPair: "竹托弹板-竹箭匣壁（诸葛弩连杆匣）", Condition: FrictionDryClean,
			MeanCoeff: 0.14, StdDev: 0.025, Low95CI: 0.10, High95CI: 0.18,
			Source: "BAOR-2019-041 古兵器摩擦学测试", MeasurementYear: 2019, SampleCount: 40,
			Method: "斜面起滑法 + 往复滑动摩擦力仪", Notes: "楚地毛竹，含水率8%"},
		{MaterialPair: "竹托弹板-竹箭匣壁", Condition: FrictionHumid90RH,
			MeanCoeff: 0.22, StdDev: 0.040, Low95CI: 0.16, High95CI: 0.29,
			Source: "BAOR-2019-042 高湿附加试验", MeasurementYear: 2019, SampleCount: 20,
			Method: "30°C/90%RH 静摩擦测量"},
		{MaterialPair: "竹托弹板-竹箭匣壁", Condition: FrictionLubricated,
			MeanCoeff: 0.07, StdDev: 0.012, Low95CI: 0.05, High95CI: 0.09,
			Source: "BAOR-2019-043 桐油润滑", MeasurementYear: 2019, SampleCount: 20,
			Method: "涂布0.1g桐油往复滑动", Notes: "军器监传统工艺"},
		{MaterialPair: "铁木复合弓床-青铜套筒（三弓弩绞车）", Condition: FrictionDryClean,
			MeanCoeff: 0.18, StdDev: 0.035, Low95CI: 0.13, High95CI: 0.24,
			Source: "NORINCO-2015-082-A8", MeasurementYear: 2015, SampleCount: 25,
			Method: "销-套摩擦转矩仪"},
		{MaterialPair: "铁木复合弓床-青铜套筒", Condition: FrictionDryDusty,
			MeanCoeff: 0.28, StdDev: 0.060, Low95CI: 0.19, High95CI: 0.38,
			Source: "NORINCO-2015-082-A9 风沙模拟", MeasurementYear: 2015, SampleCount: 15,
			Method: "ISO12103-1 A4尘注入"},
		{MaterialPair: "青铜弩机销-青铜弩机钩牙（秦弩机）", Condition: FrictionDryClean,
			MeanCoeff: 0.15, StdDev: 0.022, Low95CI: 0.12, High95CI: 0.18,
			Source: "TH-2013-QN02", MeasurementYear: 2013, SampleCount: 30,
			Method: "望山-钩牙转矩换算"},
		{MaterialPair: "青铜弩机销-青铜弩机钩牙", Condition: FrictionLowTempMinus20,
			MeanCoeff: 0.17, StdDev: 0.030, Low95CI: 0.13, High95CI: 0.22,
			Source: "TH-2013-QN03 低温试验", MeasurementYear: 2013, SampleCount: 15,
			Method: "-20°C恒温1h加载"},
		{MaterialPair: "钢-钢PVD涂层(现代参照)", Condition: FrictionLubricated,
			MeanCoeff: 0.08, StdDev: 0.010, Low95CI: 0.06, High95CI: 0.10,
			Source: "SAE J2954-2015", MeasurementYear: 2015, SampleCount: 100,
			Method: "Falex销盘/MIL润滑脂"},
	}
}

func LookupFriction(substr string, cond FrictionCondition) (float64, *FrictionMeasurement) {
	db := GetFrictionCoefficientDatabase()
	for i := range db {
		m := &db[i]
		if m.Condition != cond {
			continue
		}
		if substr == "" {
			return m.MeanCoeff, m
		}
		found := false
		for s := 0; s+len(substr) <= len(m.MaterialPair); s++ {
			if m.MaterialPair[s:s+len(substr)] == substr {
				found = true
				break
			}
		}
		if found {
			return m.MeanCoeff, m
		}
	}
	if len(db) > 0 {
		return db[0].MeanCoeff, &db[0]
	}
	return 0.12, nil
}

// ---- 数学函数 ----

func WeibullCDF(x, k, lambda float64) float64 {
	if x <= 0 { return 0 }
	return 1.0 - math.Exp(-math.Pow(x/lambda, k))
}

func WeibullQuantile(p, k, lambda float64) float64 {
	if p <= 0 { return 0 }
	if p >= 1 { return math.Inf(1) }
	return lambda * math.Pow(-math.Log(1.0-p), 1.0/k)
}

// ---- FMEA ----

func BuildFMEAMatrix() []FMEAEntry {
	es := []FMEAEntry{
		{Mode: JamDoubleFeed, Severity: 6, Occurrence: 5, Detection: 4, Description: "双进弹"},
		{Mode: JamMisfeed, Severity: 4, Occurrence: 6, Detection: 3, Description: "不送弹"},
		{Mode: JamStovepipe, Severity: 3, Occurrence: 4, Detection: 2, Description: "卡壳/烟囱"},
		{Mode: JamFollowerBind, Severity: 5, Occurrence: 3, Detection: 5, Description: "托弹板卡滞"},
		{Mode: JamSpringFatigue, Severity: 7, Occurrence: 4, Detection: 6, Description: "弹簧疲劳"},
		{Mode: JamMagazineDamage, Severity: 9, Occurrence: 2, Detection: 3, Description: "弹匣损坏"},
		{Mode: JamForeignObject, Severity: 5, Occurrence: 3, Detection: 7, Description: "异物"},
	}
	for i := range es { es[i].RPN = es[i].Severity * es[i].Occurrence * es[i].Detection }
	return es
}

// ---- Analyzer ----

type Analyzer struct {
	mu     sync.RWMutex
	params MagazineParams
	rand   *rand.Rand
}

func NewAnalyzer(params MagazineParams) *Analyzer {
	if params.BaseJamRate <= 0 { params.BaseJamRate = 1.0 / 5000 }
	return &Analyzer{params: params, rand: rand.New(rand.NewSource(time.Now().UnixNano()))}
}

func BuildParamsFromVariant(variant *model.CrossbowVariant) MagazineParams {
	capacity := 1
	if variant != nil { capacity = variant.Performance.MagazineSize }
	baseJam := 1.0 / 5000.0
	if capacity >= 10 { baseJam = 1.0 / 3500.0 } else if capacity >= 5 { baseJam = 1.0 / 4500.0 }

	friction := 0.12
	var fricRef *FrictionMeasurement
	cond := FrictionDryClean

	if variant != nil && variant.MechanismParams != nil && variant.MechanismParams.FrictionCoeff != nil {
		friction = variant.MechanismParams.FrictionCoeff.Value
		var hint string
		if variant != nil {
			switch variant.VariantCode {
			case "zhuge": hint = "竹"
			case "san-gong": hint = "铁木"
			case "bi-zhang": hint = "青铜弩机"
			}
		}
		_, ref := LookupFriction(hint, FrictionDryClean)
		fricRef = ref
	} else {
		switch {
		case variant == nil: friction, fricRef = LookupFriction("竹", FrictionDryClean)
		case variant.VariantCode == "zhuge": friction, fricRef = LookupFriction("竹", FrictionDryClean)
		case variant.VariantCode == "san-gong": friction, fricRef = LookupFriction("铁木", FrictionDryClean)
		case variant.VariantCode == "bi-zhang": friction, fricRef = LookupFriction("青铜弩机", FrictionDryClean)
		default: friction, fricRef = LookupFriction("", FrictionDryClean)
		}
	}

	muFactor := 1.0
	if friction > 0.08 {
		muFactor = (friction - 0.08) / 0.12
		if muFactor < 1 { muFactor = 1 }
		if muFactor > 3 { muFactor = 3 }
	}
	baseJam *= muFactor

	return MagazineParams{
		Capacity:          capacity,
		SpringRate:        850,
		FollowerFriction:  friction,
		FrictionSource:    fricRef,
		ToleranceClass:    "standard",
		BaseJamRate:       baseJam,
		EnvironmentalCond: cond,
	}
}

func getModeBaseWeights() map[JamFailureMode]float64 {
	return map[JamFailureMode]float64{
		JamDoubleFeed: 0.22, JamMisfeed: 0.28, JamStovepipe: 0.15, JamFollowerBind: 0.12,
		JamSpringFatigue: 0.10, JamMagazineDamage: 0.05, JamForeignObject: 0.08,
	}
}

func (a *Analyzer) Analyze(shots int, simTimeSec float64) *MagazineReliabilityAnalysis {
	a.mu.Lock()
	defer a.mu.Unlock()
	if shots <= 0 { shots = 1000 }
	if simTimeSec <= 0 { simTimeSec = 3600 }

	capacity := a.params.Capacity
	if capacity <= 0 { capacity = 10 }

	k := 2.5
	lambda := float64(capacity) * 500.0

	totalShots := shots
	jamCount := 0
	firstJamShot := WeibullQuantile(a.rand.Float64(), k, lambda)

	distribution := make(map[string]int)
	rateMap := make(map[string]float64)
	modeWeights := getModeBaseWeights()
	events := make([]JamEvent, 0)

	for s := 1; s <= totalShots; s++ {
		fatigue := float64(s) / (lambda * 2)
		if fatigue > 1 { fatigue = 1 }
		fatigueFactor := 1.0
		if fatigue > 0.8 { fatigueFactor = 10.0 } else if fatigue > 0.6 { fatigueFactor = 3.0 }

		jamProb := a.params.BaseJamRate * fatigueFactor
		if s < int(firstJamShot) {
			cdf := WeibullCDF(float64(s), k, lambda)
			jamProb *= cdf / (math.SmallestNonzeroFloat64 + cdf)
			if jamProb < a.params.BaseJamRate*0.1 { jamProb = a.params.BaseJamRate * 0.1 }
		}

		if a.rand.Float64() < jamProb {
			jamCount++
			totalW := 0.0
			for _, w := range modeWeights { totalW += w }
			r := a.rand.Float64() * totalW
			acc := 0.0
			picked := JamMisfeed
			for mode, w := range modeWeights {
				acc += w
				if r <= acc { picked = mode; break }
			}
			sev := 3.0 + a.rand.Float64()*6
			recoverable := picked != JamMagazineDamage
			recovSec := 2.0 + a.rand.Float64()*8
			if !recoverable { recovSec = 30 + a.rand.Float64()*60 }
			if picked == JamSpringFatigue { modeWeights[picked] *= 1.05 }

			distribution[string(picked)]++
			events = append(events, JamEvent{
				Mode: picked, Timestamp: int64(float64(s) / 3.0), Severity: sev,
				Recoverable: recoverable, RootCause: string(picked), RecoveryTimeSec: recovSec,
			})
		}
	}

	jamProbPerShot := 0.0
	mtbfShots := 0.0
	if jamCount > 0 {
		jamProbPerShot = float64(jamCount) / float64(totalShots)
		mtbfShots = float64(totalShots) / float64(jamCount)
	} else {
		jamProbPerShot = a.params.BaseJamRate
		mtbfShots = 1.0 / a.params.BaseJamRate
	}

	avgFireRateSec := simTimeSec / float64(totalShots)
	if avgFireRateSec <= 0 { avgFireRateSec = 6.0 }
	mtbfHours := (mtbfShots * avgFireRateSec) / 3600.0

	z95 := 1.96
	se := math.Sqrt(jamProbPerShot * (1 - jamProbPerShot) / float64(totalShots))
	mtbfLow := mtbfShots / (1 + z95*se*math.Sqrt(float64(jamCount)))
	mtbfHigh := mtbfShots / math.Max(0.0001, 1-z95*se*math.Sqrt(float64(jamCount)))

	curve := make([]model.Point, 100)
	maxShots := mtbfShots * 3
	if maxShots < 100 { maxShots = 100 }
	for i := 0; i < 100; i++ {
		n := float64(i+1) * maxShots / 100.0
		curve[i] = model.Point{X: n, Y: math.Exp(-n / mtbfShots)}
	}

	for mode, cnt := range distribution { rateMap[mode] = float64(cnt) / float64(totalShots) }
	allModes := []JamFailureMode{JamDoubleFeed, JamMisfeed, JamStovepipe, JamFollowerBind, JamSpringFatigue, JamMagazineDamage, JamForeignObject}
	for _, m := range allModes {
		if _, ok := distribution[string(m)]; !ok { distribution[string(m)] = 0; rateMap[string(m)] = 0 }
	}

	return &MagazineReliabilityAnalysis{
		TotalShots: totalShots, JamCount: jamCount, JamProbabilityPerShot: jamProbPerShot,
		MTBFShots: mtbfShots, MTBFHours: mtbfHours,
		FailureModeDistribution: distribution, FailureModeRate: rateMap,
		ReliabilityCurvePts: curve, ConfidenceInterval: ConfidenceInterval{Low: mtbfLow, High: mtbfHigh, Level: 0.95},
		JamEvents: events, FMEAMatrix: BuildFMEAMatrix(),
		FrictionCoeffUsed: a.params.FollowerFriction, FrictionMeasurementRef: a.params.FrictionSource,
		EnvironmentalCondition: a.params.EnvironmentalCond, WeibullShapeK: k, WeibullScaleLambda: lambda,
	}
}

// ---- 高级报告聚合函数 ----

func (a *Analyzer) GenerateReport(variantCode string, analysis *MagazineReliabilityAnalysis) *ReliabilityReport {
	cap := a.params.Capacity
	rating := "low"
	if cap >= 10 { rating = "high" } else if cap >= 5 { rating = "medium" }
	baseJam := a.params.BaseJamRate
	muAmplif := 1.0
	if baseJam != 0 {
		nominal := 1.0 / 5000.0
		if cap >= 10 { nominal = 1.0 / 3500.0 } else if cap >= 5 { nominal = 1.0 / 4500.0 }
		muAmplif = baseJam / nominal
	}
	mtbf := analysis.MTBFShots
	ciLowRatio := analysis.ConfidenceInterval.Low / math.Max(1e-6, mtbf)
	ciHighRatio := analysis.ConfidenceInterval.High / math.Max(1e-6, mtbf)
	occMax := JamMisfeed
	occMaxVal := -1
	sevMax := JamMagazineDamage
	sevMaxVal := -1
	rpnWorst := 0
	for _, e := range analysis.FMEAMatrix {
		if int(e.Occurrence) > occMaxVal { occMaxVal = int(e.Occurrence); occMax = e.Mode }
		if int(e.Severity) > sevMaxVal { sevMaxVal = int(e.Severity); sevMax = e.Mode }
		if e.RPN > rpnWorst { rpnWorst = e.RPN }
	}
	rAtMTBF := 0.3679
	if len(analysis.ReliabilityCurvePts) > 0 {
		for _, p := range analysis.ReliabilityCurvePts {
			if math.Abs(p.X-mtbf) < math.Abs(p.X-analysis.ReliabilityCurvePts[0].X) || rAtMTBF == 0.3679 {
				if math.Abs(p.X-mtbf) < 0.3*mtbf {
					rAtMTBF = p.Y
				}
			}
		}
	}
	curveDecay := true
	for i := 1; i < len(analysis.ReliabilityCurvePts); i++ {
		if analysis.ReliabilityCurvePts[i].Y > analysis.ReliabilityCurvePts[i-1].Y+1e-9 {
			curveDecay = false
			break
		}
	}
	return &ReliabilityReport{
		VariantCode:       variantCode,
		CapacityJamRating: rating,
		MuJamAmplification: muAmplif,
		R95LowRatio:       ciLowRatio,
		R95HighRatio:      ciHighRatio,
		ModeMaxOccurrence: occMax,
		SeverityMaxMode:   sevMax,
		RPNWorst:          rpnWorst,
		ReliabilityAtMTBF: rAtMTBF,
		CurveStrictDecay:  curveDecay,
	}
}
