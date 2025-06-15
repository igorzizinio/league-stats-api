package util

type LeagueRegion string
type RiotRegion string

const (
	RiotRegionAmericas RiotRegion = "americas"
	RiotRegionEurope   RiotRegion = "europe"
	RiotRegionAsia     RiotRegion = "asia"
	RiotRegionPBE      RiotRegion = "pbe"
	RiotRegionSEA      RiotRegion = "sea"
	RiotRegionUnknown  RiotRegion = "unknown"
)

func RiotRegionFromLeague(region string) RiotRegion {
	switch region {
	case "br1", "na1", "la1", "la2":
		return RiotRegionAmericas
	case "eun1", "euw1", "tr1", "ru":
		return RiotRegionEurope
	case "jp1", "kr":
		return RiotRegionAsia
	case "sg2", "ph2", "tw2", "th2", "vn2", "id1":
		return RiotRegionSEA
	case "pbe1":
		return RiotRegionPBE
	default:
		return RiotRegionUnknown
	}
}
