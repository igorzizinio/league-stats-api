package model

type AccountData struct {
	Puuid    string `json:"puuid"`
	GameName string `json:"gameName"`
	TagLine  string `json:"tagLine"`
}

type SummonerData struct {
	Id            string `json:"id"`
	Puuid         string `json:"puuid"`
	AccountId     string `json:"accountId"`
	ProfileIconId int    `json:"profileIconId"`
	RevisionDate  int64  `json:"revisionDate"`
	SummonerLevel int    `json:"summonerLevel"`
}

type MatchData struct {
	Info struct {
		Participants []MatchParticipant `json:"participants"`
	} `json:"info"`
}

type TimelineData struct {
	Info struct {
		Frames []struct {
			Events []MatchEvent `json:"events"`
		} `json:"frames"`
	} `json:"info"`
}

type MatchParticipant struct {
	Puuid                       string `json:"puuid"`
	RiotIdGameName              string `json:"riotIdGameName"`
	RiotIdTagline               string `json:"riotIdTagline"`
	ChampionName                string `json:"championName"`
	TeamPosition                string `json:"teamPosition"`
	Role                        string `json:"role"`
	Kills                       int    `json:"kills"`
	Deaths                      int    `json:"deaths"`
	Assists                     int    `json:"assists"`
	TotalDamageDealtToChampions int    `json:"totalDamageDealtToChampions"`
	VisionScore                 int    `json:"visionScore"`
	Item0                       int    `json:"item0"`
	Item1                       int    `json:"item1"`
	Item2                       int    `json:"item2"`
	Item3                       int    `json:"item3"`
	Item4                       int    `json:"item4"`
	Item5                       int    `json:"item5"`
	Item6                       int    `json:"item6"`
	ParticipantId               int    `json:"participantId"`
}

type MatchEvent struct {
	Type          string
	Timestamp     int
	ParticipantId int
	KillerId      int
	VictimId      int
	ItemId        string
	SkillSlot     int
}

type ItemData struct {
	Name string
}

type OptimizedMatch struct {
	Name         string
	Champion     string
	Role         string
	KDA          string
	TotalDamage  int
	VisionScore  int
	Items        []int
	EventSummary []string
}

type DDragonItemData struct {
	Data map[string]ItemData `json:"data"`
}
