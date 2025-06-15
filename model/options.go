package model

type GetMatchesOptions struct {
	StartTime  *int64  `json:"startTime,omitempty" query:"startTime,omitempty" form:"startTime,omitempty"`
	EndTime    *int64  `json:"endTime,omitempty" query:"endTime,omitempty" form:"endTime,omitempty"`
	Queue      *int    `json:"queue,omitempty" query:"queue,omitempty" form:"queue,omitempty"`
	Type       *string `json:"type,omitempty" query:"type,omitempty" form:"type,omitempty"`
	StartIndex *int    `json:"start,omitempty" query:"start,omitempty" form:"start,omitempty"`
	Count      *int    `json:"count,omitempty" query:"count,omitempty" form:"count,omitempty"`
}

/*
	startTime?: number
  	endTime?: number
  	queue?: number
  	type?: string
  	start?: number
  	count?: number
*/
