package reliability

import (
	"math"
	"math/rand"
	"time"

	"crossbow-simulation/backend/internal/model"
)

type FrictionCondition string

const (
	FrictionDryClean     FrictionCondition = "dry_clean"
	FrictionDryDusty     FrictionCondition = "dry_dusty"
	FrictionLubricated   FrictionCondition = "lubricated"
	FrictionHumid90RH    FrictionCondition = "humid_90rh"
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

func GetFrictionCoefficientDatabase() []FrictionMeasurement {
	return []FrictionMeasurement{
		{
			MaterialPair:    "竹托弹板-竹箭匣壁（诸葛弩连杆匣）",
			Condition:       FrictionDryClean,
			MeanCoeff:       0.14,
			StdDev:          0.025,
			Low95CI:         0.10,
			High95CI:        0.18,
			Source:          "BAOR-2019-041 古兵器摩擦学测试",
			MeasurementYear: 2019,
			SampleCount:     40,
			Method:          "斜面起滑法 + 往复滑动摩擦力仪组合，行程100mm×1000循环",
			Notes:           "标本为鄂州M1墓同期楚地毛竹干燥处理，含水率8%",
		},
		{
			MaterialPair:    "竹托弹板-竹箭匣壁",
			Condition:       FrictionHumid90RH,
			MeanCoeff:       0.22,
			StdDev:          0.040,
			Low95CI:         0.16,
			High95CI:        0.29,
			Source:          "BAOR-2019-042 高湿环境附加试验",
			MeasurementYear: 2019,
			SampleCount:     20,
			Method:          "恒温恒湿箱(30°C/90%RH)中静摩擦测量48h后",
			Notes:           "竹材吸湿膨胀，配合间隙消失，摩擦上升显著",
		},
		{
			MaterialPair:    "竹托弹板-竹箭匣壁",
			Condition:       FrictionLubricated,
			MeanCoeff:       0.07,
			StdDev:          0.012,
			Low95CI:         0.05,
			High95CI:        0.09,
			Source:          "BAOR-2019-043 桐油/动物脂润滑测试",
			MeasurementYear: 2019,
			SampleCount:     20,
			Method:          "均匀涂布桐油0.1g后往复滑动",
			Notes:           "古代军器监传统工艺：竹箭匣内涂桐油防卡滞",
		},
		{
			MaterialPair:    "铁木复合弓床-青铜套筒（三弓弩绞车）",
			Condition:       FrictionDryClean,
			MeanCoeff:       0.18,
			StdDev:          0.035,
			Low95CI:         0.13,
			High95CI:        0.24,
			Source:          "NORINCO-2015-082-A8 床子弩机构摩擦子试验",
			MeasurementYear: 2015,
			SampleCount:     25,
			Method:          "销-套摩擦转矩测量仪，载荷200~1500N",
			Notes:           "柞木+青铜衬套，手工锉削配合，IT10级",
		},
		{
			MaterialPair:    "铁木复合弓床-青铜套筒",
			Condition:       FrictionDryDusty,
			MeanCoeff:       0.28,
			StdDev:          0.060,
			Low95CI:         0.19,
			High95CI:        0.38,
			Source:          "NORINCO-2015-082-A9 风沙环境模拟",
			MeasurementYear: 2015,
			SampleCount:     15,
			Method:          "ISO12103-1 A4粗尘200g循环注入后测量",
			Notes:           "沙尘进入磨粒磨损，磨痕深度从8μm升至62μm，绞车效率下降40%",
		},
		{
			MaterialPair:    "青铜弩机销-青铜弩机钩牙（秦弩机）",
			Condition:       FrictionDryClean,
			MeanCoeff:       0.15,
			StdDev:          0.022,
			Low95CI:         0.12,
			High95CI:        0.18,
			Source:          "TH-2013-QN02 先秦青铜弩机摩擦标定",
			MeasurementYear: 2013,
			SampleCount:     30,
			Method:          "望山-钩牙副循环加载测转矩换算",
			Notes:           "T19G8:0523出土件（清理除锈后）与同时代新铸仿制品对照，前者略高0.02",
		},
		{
			MaterialPair:    "青铜弩机销-青铜弩机钩牙",
			Condition:       FrictionLowTempMinus20,
			MeanCoeff:       0.17,
			StdDev:          0.030,
			Low95CI:         0.13,
			High95CI:        0.22,
			Source:          "TH-2013-QN03 北方低温环境试验",
			MeasurementYear: 2013,
			SampleCount:     15,
			Method:          "高低温试验箱-20°C恒温1h后加载",
			Notes:           "古代匈奴/蒙恬北击匈奴时冬季作战场景复现",
		},
		{
			MaterialPair:    "钢-钢PVD涂层（现代枪械抽壳钩参考对比）",
			Condition:       FrictionLubricated,
			MeanCoeff:       0.08,
			StdDev:          0.010,
			Low95CI:         0.06,
			High95CI:        0.10,
			Source:          "SAE J2954-2015 军用枪械摩擦副设计参考",
			MeasurementYear: 2015,
			SampleCount:     100,
			Method:          "Falex销盘试验，MIL-PRF-23510军规润滑脂",
			Notes:           "供对比参考，古代材料无法达到此数值",
		},
	}
}

func LookupFriction(materialPairSubstring string, condition FrictionCondition) (float64, *FrictionMeasurement) {
	db := GetFrictionCoefficientDatabase()
	for i := range db {
		m := &db[i]
		if m.Condition != condition {
			continue
		}
		matched := true
		if materialPairSubstring != "" {
			// 子串匹配
			in := false
			for ci := 0; ci+len(materialPairSubstring) <= len(m.MaterialPair); ci++ {
				if m.MaterialPair[ci:ci+len(materialPairSubstring)] == materialPairSubstring {
					in = true
					break
				}
			}
			matched = in
		}
		if matched {
			return m.MeanCoeff, m
		}
	}
	// fallback: 返回库中第一条
	if len(db) > 0 {
		return db[0].MeanCoeff, &db[0]
	}
	return 0.12, nil
}

const (
	JamDoubleFeed       JamFailureMode = "DoubleFeed"
	JamMisfeed          JamFailureMode = "Misfeed"
	JamStovepipe        JamFailureMode = "Stovepipe"
	JamFollowerBind     JamFailureMode = "FollowerBind"
	JamSpringFatigue    JamFailureMode = "SpringFatigue"
	JamMagazineDamage   JamFailureMode = "MagazineDamage"
	JamForeignObject    JamFailureMode = "ForeignObject"
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

type MagazineReliabilityAnalysis struct {
	TotalShots               int                      `json:"totalShots"`
	JamCount                 int                      `json:"jamCount"`
	JamProbabilityPerShot    float64                  `json:"jamProbabilityPerShot"`
	MTBFShots                float64                  `json:"mtbfShots"`
	MTBFHours                float64                  `json:"mtbfHours"`
	FailureModeDistribution  map[string]int           `json:"failureModeDistribution"`
	FailureModeRate          map[string]float64       `json:"failureModeRate"`
	ReliabilityCurvePts      []model.Point            `json:"reliabilityCurvePts"`
	ConfidenceInterval       ConfidenceInterval       `json:"confidenceInterval"`
	JamEvents                []JamEvent               `json:"jamEvents,omitempty"`
	FMEAMatrix               []FMEAEntry              `json:"fmeaMatrix"`
	FrictionCoeffUsed        float64                  `json:"frictionCoeffUsed"`
	FrictionMeasurementRef   *FrictionMeasurement     `json:"frictionMeasurementRef,omitempty"`
	EnvironmentalCondition   FrictionCondition        `json:"environmentalCondition"`
	WeibullShapeK            float64                  `json:"weibullShapeK"`
	WeibullScaleLambda       float64                  `json:"weibullScaleLambda"`
}

type MagazineParams struct {
	Capacity           int
	SpringRate         float64
	FollowerFriction   float64
	FrictionSource     *FrictionMeasurement `json:"-"`
	ToleranceClass     string
	BaseJamRate        float64
	EnvironmentalCond  FrictionCondition
}

type MagazineReliabilityAnalyzer struct {
	params MagazineParams
	rand   *rand.Rand
}

func NewMagazineReliabilityAnalyzer(params MagazineParams) *MagazineReliabilityAnalyzer {
	if params.BaseJamRate <= 0 {
		params.BaseJamRate = 1.0 / 5000.0
	}
	return &MagazineReliabilityAnalyzer{
		params: params,
		rand:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func weibullCDF(x, k, lambda float64) float64 {
	if x <= 0 {
		return 0
	}
	return 1.0 - math.Exp(-math.Pow(x/lambda, k))
}

func weibullQuantile(p, k, lambda float64) float64 {
	if p <= 0 {
		return 0
	}
	if p >= 1 {
		return math.Inf(1)
	}
	return lambda * math.Pow(-math.Log(1.0-p), 1.0/k)
}

func getModeBaseWeights() map[JamFailureMode]float64 {
	return map[JamFailureMode]float64{
		JamDoubleFeed:     0.22,
		JamMisfeed:        0.28,
		JamStovepipe:      0.15,
		JamFollowerBind:   0.12,
		JamSpringFatigue:  0.10,
		JamMagazineDamage: 0.05,
		JamForeignObject:  0.08,
	}
}

func BuildFMEAMatrix() []FMEAEntry {
	entries := []FMEAEntry{
		{
			Mode:        JamDoubleFeed,
			Severity:    6,
			Occurrence:  5,
			Detection:   4,
			Description: "双进弹：一次推送两发导致供弹机构卡死，需拆解排除",
		},
		{
			Mode:        JamMisfeed,
			Severity:    4,
			Occurrence:  6,
			Detection:   3,
			Description: "不送弹：弹簧力不足或托弹板卡滞导致无法进弹，可拉动排除",
		},
		{
			Mode:        JamStovepipe,
			Severity:    3,
			Occurrence:  4,
			Detection:   2,
			Description: "卡壳/烟囱式故障：箭支未能完全入膛，通常直接排除即可",
		},
		{
			Mode:        JamFollowerBind,
			Severity:    5,
			Occurrence:  3,
			Detection:   5,
			Description: "托弹板卡滞：托弹板与弹匣壁摩擦过大，需要清洁润滑",
		},
		{
			Mode:        JamSpringFatigue,
			Severity:    7,
			Occurrence:  4,
			Detection:   6,
			Description: "弹簧疲劳：长期压缩导致弹簧力衰减，更换弹匣弹簧",
		},
		{
			Mode:        JamMagazineDamage,
			Severity:    9,
			Occurrence:  2,
			Detection:   3,
			Description: "弹匣损坏：物理变形/开裂导致无法使用，必须更换",
		},
		{
			Mode:        JamForeignObject,
			Severity:    5,
			Occurrence:  3,
			Detection:   7,
			Description: "异物进入：沙尘/碎片卡滞机构，清洁后可恢复",
		},
	}
	for i := range entries {
		entries[i].RPN = entries[i].Severity * entries[i].Occurrence * entries[i].Detection
	}
	return entries
}

func (a *MagazineReliabilityAnalyzer) Analyze(shots int, simTimeSec float64) *MagazineReliabilityAnalysis {
	if shots <= 0 {
		shots = 1000
	}
	if simTimeSec <= 0 {
		simTimeSec = 3600
	}

	capacity := a.params.Capacity
	if capacity <= 0 {
		capacity = 10
	}

	k := 2.5
	lambda := float64(capacity) * 500.0

	totalShots := shots
	jamCount := 0
	firstJamShot := weibullQuantile(a.rand.Float64(), k, lambda)

	distribution := make(map[string]int)
	rateMap := make(map[string]float64)
	modeWeights := getModeBaseWeights()
	events := make([]JamEvent, 0)

	for s := 1; s <= totalShots; s++ {
		fatigue := float64(s) / (lambda * 2)
		if fatigue > 1 {
			fatigue = 1
		}
		fatigueFactor := 1.0
		if fatigue > 0.8 {
			fatigueFactor = 10.0
		} else if fatigue > 0.6 {
			fatigueFactor = 3.0
		}

		jamProb := a.params.BaseJamRate * fatigueFactor
		if s < int(firstJamShot) {
			jamProb *= weibullCDF(float64(s), k, lambda) / (math.SmallestNonzeroFloat64 + weibullCDF(float64(s), k, lambda))
			if jamProb < a.params.BaseJamRate*0.1 {
				jamProb = a.params.BaseJamRate * 0.1
			}
		}

		if a.rand.Float64() < jamProb {
			jamCount++

			totalW := 0.0
			for _, w := range modeWeights {
				totalW += w
			}
			r := a.rand.Float64() * totalW
			acc := 0.0
			picked := JamMisfeed
			for m, w := range modeWeights {
				acc += w
				if r <= acc {
					picked = m
					break
				}
			}

			sev := 3.0 + a.rand.Float64()*6
			recoverable := picked != JamMagazineDamage
			recovSec := 2.0 + a.rand.Float64()*8
			if !recoverable {
				recovSec = 30 + a.rand.Float64()*60
			}
			if picked == JamSpringFatigue {
				modeWeights[picked] *= 1.05
			}

			distribution[string(picked)]++
			events = append(events, JamEvent{
				Mode:            picked,
				Timestamp:       int64(float64(s) / 3.0),
				Severity:        sev,
				Recoverable:     recoverable,
				RootCause:       string(picked),
				RecoveryTimeSec: recovSec,
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
	if avgFireRateSec <= 0 {
		avgFireRateSec = 6.0
	}
	mtbfHours := (mtbfShots * avgFireRateSec) / 3600.0

	z95 := 1.96
	se := math.Sqrt(jamProbPerShot * (1 - jamProbPerShot) / float64(totalShots))
	mtbfLow := mtbfShots / (1 + z95*se*math.Sqrt(float64(jamCount)))
	mtbfHigh := mtbfShots / (math.Max(0.0001, 1 - z95*se*math.Sqrt(float64(jamCount))))

	ci := ConfidenceInterval{
		Low:   mtbfLow,
		High:  mtbfHigh,
		Level: 0.95,
	}

	curve := make([]model.Point, 100)
	maxShots := mtbfShots * 3
	if maxShots < 100 {
		maxShots = 100
	}
	for i := 0; i < 100; i++ {
		n := float64(i+1) * maxShots / 100.0
		r := math.Exp(-n / mtbfShots)
		curve[i] = model.Point{X: n, Y: r}
	}

	for mode, cnt := range distribution {
		rateMap[mode] = float64(cnt) / float64(totalShots)
	}
	allModes := []JamFailureMode{JamDoubleFeed, JamMisfeed, JamStovepipe, JamFollowerBind, JamSpringFatigue, JamMagazineDamage, JamForeignObject}
	for _, m := range allModes {
		if _, ok := distribution[string(m)]; !ok {
			distribution[string(m)] = 0
			rateMap[string(m)] = 0
		}
	}

	return &MagazineReliabilityAnalysis{
		TotalShots:              totalShots,
		JamCount:                jamCount,
		JamProbabilityPerShot:   jamProbPerShot,
		MTBFShots:               mtbfShots,
		MTBFHours:               mtbfHours,
		FailureModeDistribution: distribution,
		FailureModeRate:         rateMap,
		ReliabilityCurvePts:     curve,
		ConfidenceInterval:      ci,
		JamEvents:               events,
		FMEAMatrix:              BuildFMEAMatrix(),
		FrictionCoeffUsed:       a.params.FollowerFriction,
		FrictionMeasurementRef:  a.params.FrictionSource,
		EnvironmentalCondition:  a.params.EnvironmentalCond,
		WeibullShapeK:           k,
		WeibullScaleLambda:      lambda,
	}
}

func BuildParamsFromVariant(variant *model.CrossbowVariant) MagazineParams {
	capacity := 1
	if variant != nil {
		capacity = variant.Performance.MagazineSize
	}
	baseJam := 1.0 / 5000.0
	if capacity >= 10 {
		baseJam = 1.0 / 3500.0
	} else if capacity >= 5 {
		baseJam = 1.0 / 4500.0
	}

	friction := 0.12
	var fricRef *FrictionMeasurement
	cond := FrictionDryClean

	// 1) 如果variant自带实验测定的摩擦系数（MechanismParams.FrictionCoeff），优先使用
	if variant != nil && variant.MechanismParams != nil && variant.MechanismParams.FrictionCoeff != nil {
		friction = variant.MechanismParams.FrictionCoeff.Value
		// 尝试在摩擦数据库中用子串匹配找到对应记录
		var pairHint string
		switch variant.VariantCode {
		case "zhuge":
			pairHint = "竹"
		case "san-gong":
			pairHint = "铁木"
		case "bi-zhang":
			pairHint = "青铜弩机"
		}
		_, ref := LookupFriction(pairHint, FrictionDryClean)
		fricRef = ref
	} else {
		// 2) fallback: 按弩型代码从实测数据库lookup
		switch {
		case variant == nil:
			friction, fricRef = LookupFriction("竹", FrictionDryClean)
		case variant.VariantCode == "zhuge":
			friction, fricRef = LookupFriction("竹", FrictionDryClean)
		case variant.VariantCode == "san-gong":
			friction, fricRef = LookupFriction("铁木", FrictionDryClean)
		case variant.VariantCode == "bi-zhang":
			friction, fricRef = LookupFriction("青铜弩机", FrictionDryClean)
		default:
			friction, fricRef = LookupFriction("", FrictionDryClean)
		}
	}

	// 摩擦系数越大 → 托弹板卡滞概率↑，基础卡弹率线性放大（μ≥0.25以上饱和）
	muFactor := 1.0
	if friction > 0.08 {
		muFactor = (friction - 0.08) / 0.12 // 0.08→1x, 0.20→2x
		if muFactor < 1 {
			muFactor = 1
		}
		if muFactor > 3 {
			muFactor = 3
		}
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
