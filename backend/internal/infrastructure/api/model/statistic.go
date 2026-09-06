package model

type StatisticResponse struct {
	Body Statistic `json:"body" bson:"body"`
}

// Statistic TODO: think about adding more values like visitors, ...
type Statistic struct {
	ShelfNumber   int `json:"shelf_number"`
	SectionNumber int `json:"section_number"`
	LinkNumber    int `json:"link_number"`
}
