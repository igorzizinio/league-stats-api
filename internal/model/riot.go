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
		Frames []MatchFrame `json:"frames"`
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

	TotalDamageTaken int `json:"totalDamageTaken"`

	PhysicalDamageDealt            int `json:"physicalDamageDealt"`
	PhysicalDamageDealtToChampions int `json:"physicalDamageDealtToChampions"`

	MagicDamageDealt            int `json:"magicDamageDealt"`
	MagicDamageDealtToChampions int `json:"magicDamageDealtToChampions"`

	TrueDamageDealt            int `json:"trueDamageDealt"`
	TrueDamageDealtToChampions int `json:"trueDamageDealtToChampions"`

	TotalHeal int `json:"totalHeal"`

	DamageSelfMitigated int `json:"damageSelfMitigated"`

	TotalHealsOnTeammates   int  `json:"totalHealsOnTeammates"`
	DamageDealtToObjectives int  `json:"damageDealtToObjectives"`
	DamageDealtToTurrets    int  `json:"damageDealtToTurrets"`
	TimeCCingOthers         int  `json:"timeCCingOthers"`
	FirstBloodKill          bool `json:"firstBloodKill"`
	FirstTowerKill          bool `json:"firstTowerKill"`

	Challenges map[string]any `json:"challenges"`
}

type ParticipantPerks struct {
	StatPerks struct {
		Defense int `json:"defense"`
		Flex    int `json:"flex"`
		Offense int `json:"offense"`
	} `json:"statPerks"`
	Styles []struct {
		// TODO: documentar melhor isso, pq oq é var1, var2, var3? e o que é perk?
		Selections []struct {
			Perk int `json:"perk"`
			Var1 int `json:"var1"`
			Var2 int `json:"var2"`
			Var3 int `json:"var3"`
		} `json:"selections"`
		Description string `json:"description"`
	} `json:"styles"`
}

type MatchEvent struct {
	Type          string `json:"type"`
	Timestamp     int    `json:"timestamp"`
	RealTimestamp int    `json:"realTimestamp"`

	// Não documentado, mas podem estar presentes?
	ParticipantId           int    `json:"participantId"`
	KillerId                int    `json:"killerId"`
	VictimId                int    `json:"victimId"`
	ItemId                  int    `json:"itemId"`
	SkillSlot               int    `json:"skillSlot"`
	KillType                string `json:"killType"`
	MonsterType             string `json:"monsterType"`
	MonsterSubType          string `json:"monsterSubType"`
	BuildingType            string `json:"buildingType"`
	LaneType                string `json:"laneType"`
	WardType                string `json:"wardType"`
	AssistingParticipantIds []int  `json:"assistingParticipantIds"`
	CreatorId               int    `json:"creatorId"`
	TeamId                  int    `json:"teamId"`
	LevelUpType             string `json:"levelUpType"`
	KillStreakLength        int    `json:"killStreakLength"`
	Bounty                  int    `json:"bounty"`
}

type MatchFrame struct {
	Timestamp         int                              `json:"timestamp"`
	ParticipantFrames map[string]MatchParticipantFrame `json:"participantFrames"`
	Events            []MatchEvent                     `json:"events"`
}

type MatchParticipantFrame struct {
	ChampionStats            map[string]any     `json:"championStats"`
	CurrentGold              int                `json:"currentGold"`
	DamageStats              map[string]any     `json:"damageStats"`
	GoldPerSecond            int                `json:"goldPerSecond"`
	JungleMinionsKilled      int                `json:"jungleMinionsKilled"`
	Level                    int                `json:"level"`
	MinionsKilled            int                `json:"minionsKilled"`
	Position                 map[string]float64 `json:"position"`
	TimeEnemySpentControlled int                `json:"timeEnemySpentControlled"`
	TotalGold                int                `json:"totalGold"`
	XP                       int                `json:"xp"`
}

type ItemData struct {
	Name string
}

type OptimizedMatch struct {
	Name             string
	Champion         string
	Role             string
	KDA              string
	TotalDamage      int
	VisionScore      int
	Items            []string
	EventSummary     []string
	GameContext      map[string]any            `json:"gameContext,omitempty"`
	PlayerSnapshot   map[string]any            `json:"playerSnapshot,omitempty"`
	PhaseStats       map[string]map[string]int `json:"phaseStats,omitempty"`
	CombatStats      map[string]int            `json:"combatStats,omitempty"`
	ObjectiveStats   map[string]int            `json:"objectiveStats,omitempty"`
	VisionStats      map[string]int            `json:"visionStats,omitempty"`
	DataCompleteness map[string]int            `json:"dataCompleteness,omitempty"`
	// ItemsFull contains the full DDragon items dataset (keyed by item id)
	ItemsFull map[string]any `json:"itemsFull,omitempty"`
	// ChampionData contains the champion details (spells, scalings, stats)
	ChampionData map[string]any `json:"championData,omitempty"`
	// ItemSummaries provides per-item status and event counts for the participant
	ItemSummaries    map[string]map[string]any `json:"itemSummaries,omitempty"`
	OpponentSnapshot map[string]any            `json:"opponentSnapshot,omitempty"`
	Challenges       map[string]any            `json:"challenges,omitempty"`
	Runes            map[string]any            `json:"runes,omitempty"`
	SkillOrder       []int                     `json:"skillOrder,omitempty"`
	GoldXpTimeline   []map[string]any          `json:"goldXpTimeline,omitempty"`
}

type DDragonItemData struct {
	Data map[string]ItemData `json:"data"`
}
