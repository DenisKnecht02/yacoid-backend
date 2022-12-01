package types

type StatisticsResponse struct {
	DefinitionCount                 int `json:"definitionCount"`
	DefinitionCountInCurrentQuarter int `json:"definitionCountInCurrentQuarter"`
	SourceCount                     int `json:"sourceCount"`
	SourceCountInCurrentQuarter     int `json:"sourceCountInCurrentQuarter"`
	AuthorCount                     int `json:"authorCount"`
	AuthorCountInCurrentQuarter     int `json:"authorCountInCurrentQuarter"`
}
