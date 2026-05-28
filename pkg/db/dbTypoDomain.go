package db

import (
	"github.com/moorada/neferpitool/pkg/constants"
	"github.com/moorada/neferpitool/pkg/domains"
)

func AddTypoDomainToDB(td domains.TypoDomain) {
	db.Create(&td)
}

func GetTypoDomainFromDB(nameTypoDomain string) domains.TypoDomain {

	var typoDomain domains.TypoDomain
	db.Where("name = ?", nameTypoDomain).Last(&typoDomain)
	return typoDomain
}

func RemoveTypoDomainFromDB(nameTypoDomain string) {
	db.Where("name = ?", nameTypoDomain).Delete(&domains.TypoDomain{})
}

func GetTypoDomainHistoryFromDB(typoDomain string) domains.TypoList {
	var tds []domains.TypoDomain
	db.Where("name = ?", typoDomain).Find(&tds)
	return tds
}

// HasTypoHostsForZone reports whether typo-squat hosts were already generated for the apex.
func HasTypoHostsForZone(apex string) bool {
	var count int
	db.Model(&domains.TypoDomain{}).Where("legit_domain = ? AND source = ?", apex, constants.SourceTypo).Count(&count)
	return count > 0
}
