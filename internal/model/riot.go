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
	Metadata struct {
		MatchId string `json:"matchId"`
	} `json:"metadata"`
	Info struct {
		PlatformId   string `json:"platformId"`
		GameMode     string `json:"gameMode"`
		GameName     string `json:"gameName"`
		GameType     string `json:"gameType"`
		GameCreation int64  `json:"gameCreation"`
		GameDuration int    `json:"gameDuration"`
		QueueId      int    `json:"queueId"`
		Teams        []struct {
			TeamId int  `json:"teamId"`
			Win    bool `json:"win"`
		} `json:"teams"`
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
	RiotIdGameName string `json:"riotIdGameName"`
	RiotIdTagline  string `json:"riotIdTagline"`

	Puuid         string `json:"puuid"`
	SummonerName  string `json:"summonerName"`
	ChampionName  string `json:"championName"`
	ChampLevel    int    `json:"champLevel"`
	Role          string `json:"role"`
	TeamId        int    `json:"teamId"`
	TeamPosition  string `json:"teamPosition"`
	ParticipantId int    `json:"participantId"`

	Assists int `json:"assists"`
	Deaths  int `json:"deaths"`
	Kills   int `json:"kills"`

	Item0          int `json:"item0"`
	Item1          int `json:"item1"`
	Item2          int `json:"item2"`
	Item3          int `json:"item3"`
	Item4          int `json:"item4"`
	Item5          int `json:"item5"`
	Item6          int `json:"item6"`
	ItemsPurchased int `json:"itemsPurchased"`

	Summoner1Id int `json:"summoner1Id"`
	Summoner2Id int `json:"summoner2Id"`

	Win bool `json:"win"`

	Perks ParticipantPerks `json:"perks"`

	TotalMinionsKilled   int `json:"totalMinionsKilled"`
	NeutralMinionsKilled int `json:"neutralMinionsKilled"`
	VisionScore          int `json:"visionScore"`

	GoldEarned int `json:"goldEarned"`
	GoldSpent  int `json:"goldSpent"`

	TotalDamageDealt            int `json:"totalDamageDealt"`
	TotalDamageDealtToChampions int `json:"totalDamageDealtToChampions"`

	PhysicalDamageDealt            int `json:"physicalDamageDealt"`
	PhysicalDamageDealtToChampions int `json:"physicalDamageDealtToChampions"`

	MagicDamageDealt            int `json:"magicDamageDealt"`
	MagicDamageDealtToChampions int `json:"magicDamageDealtToChampions"`

	TrueDamageDealt            int `json:"trueDamageDealt"`
	TrueDamageDealtToChampions int `json:"trueDamageDealtToChampions"`
}

type ParticipantPerks = map[string]any

type MatchEvent struct {
	Type          string
	Timestamp     int
	ParticipantId int
	KillerId      int
	VictimId      int
	ItemId        string
	SkillSlot     int
	KillType      string
	MonsterType   string
	BuildingType  string
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
	Items        []string
	EventSummary []string
}

type DDragonItemData struct {
	Data map[string]ItemData `json:"data"`
}
