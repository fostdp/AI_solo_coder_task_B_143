package reliability

import (
	"math"
	"math/rand"
	"time"

	"crossbow-simulation/backend/internal/model"
)

type JamFailureMode string

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
}

type MagazineParams struct {
	Capacity          int
	SpringRate        float64
	FollowerFriction  float64
	ToleranceClass    string
	BaseJamRate       float64
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
	return MagazineParams{
		Capacity:         capacity,
		SpringRate:       850,
		FollowerFriction: 0.12,
		ToleranceClass:   "standard",
		BaseJamRate:      baseJam,
	}
}
