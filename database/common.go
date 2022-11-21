package database

import "yacoid_server/types"

func GetStatistics() (*types.StatisticsResponse, error) {

	definitionCount, err := GetDefinitionCount()

	if err != nil {
		return nil, err
	}

	definitionInCurrentQuarter, err := GetDefinitionCountInCurrentQuarter()

	if err != nil {
		return nil, err
	}

	sourceCount, err := GetSourceCount()

	if err != nil {
		return nil, err
	}

	sourceInCurrentQuarter, err := GetSourceCountInCurrentQuarter()

	if err != nil {
		return nil, err
	}

	authorCount, err := GetAuthorCount()

	if err != nil {
		return nil, err
	}

	authorInCurrentQuarter, err := GetAuthorCountInCurrentQuarter()

	if err != nil {
		return nil, err
	}

	response := types.StatisticsResponse{
		DefinitionCount:                 int(definitionCount),
		DefinitionCountInCurrentQuarter: int(definitionInCurrentQuarter),
		SourceCount:                     int(sourceCount),
		SourceCountInCurrentQuarter:     int(sourceInCurrentQuarter),
		AuthorCount:                     int(authorCount),
		AuthorCountInCurrentQuarter:     int(authorInCurrentQuarter),
	}

	return &response, nil

}
