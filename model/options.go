package model

type GetMatchesOptions struct {
	StartTime  *int64  `json:"startTime,omitempty"`
	EndTime    *int64  `json:"endTime,omitempty"`
	Queue      *int    `json:"queue,omitempty"`
	Type       *string `json:"type,omitempty"`
	StartIndex *int    `json:"start,omitempty"`
	Count      *int    `json:"count,omitempty"`
}

/*
	startTime?: number
  	endTime?: number
  	queue?: number
  	type?: string
  	start?: number
  	count?: number
*/
