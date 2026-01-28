package main

import (
	"errors"
	"fmt"
	"git.sr.ht/~hannes/clonya/common"
	"git.sr.ht/~hannes/laser"
	"os"
	"strconv"
	"time"
)

type database struct {
	searchCriteria common.SearchCriteria
	full           bool
	repositories   []common.Repository
}

func ReadDatabase(path string) (database, error) {
	db := database{}

	input, err := os.Open(path)
	if err != nil {
		return db, nil
	}
	defer input.Close()

	parsed, err := laser.Read(input)
	if err != nil {
		return db, err
	}

	searchCriteriaFound := false
	db.repositories = make([]common.Repository, 0, len(parsed)-1)
	for _, node := range parsed {
		l := node.AsList()
		switch l[0].AsAtom() {
		case "searchCriteria":
			err = parseSearchCriteria(&db, l)
			if err != nil {
				return db, err
			}
			searchCriteriaFound = true
		case "repository":
			id := l[1].AsAtom()
			hash := l[2].AsAtom()
			db.repositories = append(db.repositories, common.Repository{Id: string(id), CommitHash: string(hash)})
		case "full":
			db.full = true
		}
	}

	if !searchCriteriaFound {
		return db, errors.New("database does not contain a searchCriteria node")
	}

	return db, nil
}

func parseSearchCriteria(db *database, node laser.List) error {
	searchCriteriaList := node[1:]
	for _, criteria := range searchCriteriaList {
		cList := criteria.AsList()
		propName := cList[0].AsAtom()
		propValue := string(cList[1].AsAtom())
		err := error(nil)
		switch propName {
		case "allowArchived":
			db.searchCriteria.AllowArchived, err = strconv.ParseBool(propValue)
		case "allowForks":
			db.searchCriteria.AllowForks, err = strconv.ParseBool(propValue)
		case "forge":
			db.searchCriteria.Forge = common.Forge(propValue)
		case "lang":
			db.searchCriteria.Language = propValue
		case "limit":
			db.searchCriteria.Limit, err = strconv.Atoi(propValue)
		case "minCreateDate":
			db.searchCriteria.MinCreateDate, err = time.Parse(time.DateOnly, propValue)
		case "minPushDate":
			db.searchCriteria.MinPushDate, err = time.Parse(time.DateOnly, propValue)
		case "minStars":
			db.searchCriteria.MinStars, err = strconv.Atoi(propValue)
		case "maxCreateDate":
			db.searchCriteria.MaxCreateDate, err = time.Parse(time.DateOnly, propValue)
		case "maxPushDate":
			db.searchCriteria.MaxPushDate, err = time.Parse(time.DateOnly, propValue)
		case "maxStars":
			db.searchCriteria.MaxStars, err = strconv.Atoi(propValue)
		default:
			return errors.New("unknown search criterion " + string(propName))
		}
		if err != nil {
			return fmt.Errorf("unable to read search criterion %s: %v", propName, err)
		}
	}
	return nil
}

func WriteDatabase(db database, path string) error {
	output, err := os.Create(path)
	if err != nil {
		return err
	}
	defer output.Close()

	searchCriteriaNode := laser.List{laser.Atom("searchCriteria")}
	searchCriteriaNode = append(searchCriteriaNode, laser.List{laser.Atom("allowArchived"), laser.Atom(strconv.FormatBool(db.searchCriteria.AllowArchived))})
	searchCriteriaNode = append(searchCriteriaNode, laser.List{laser.Atom("allowForks"), laser.Atom(strconv.FormatBool(db.searchCriteria.AllowForks))})
	searchCriteriaNode = append(searchCriteriaNode, laser.List{laser.Atom("forge"), laser.Atom(db.searchCriteria.Forge)})
	searchCriteriaNode = append(searchCriteriaNode, laser.List{laser.Atom("lang"), laser.Atom(db.searchCriteria.Language)})
	searchCriteriaNode = append(searchCriteriaNode, laser.List{laser.Atom("limit"), laser.Atom(strconv.Itoa(db.searchCriteria.Limit))})
	searchCriteriaNode = append(searchCriteriaNode, laser.List{laser.Atom("minCreateDate"), laser.Atom(db.searchCriteria.MinCreateDate.Format(time.DateOnly))})
	searchCriteriaNode = append(searchCriteriaNode, laser.List{laser.Atom("minPushDate"), laser.Atom(db.searchCriteria.MinPushDate.Format(time.DateOnly))})
	searchCriteriaNode = append(searchCriteriaNode, laser.List{laser.Atom("minStars"), laser.Atom(strconv.Itoa(db.searchCriteria.MinStars))})
	searchCriteriaNode = append(searchCriteriaNode, laser.List{laser.Atom("maxCreateDate"), laser.Atom(db.searchCriteria.MaxCreateDate.Format(time.DateOnly))})
	searchCriteriaNode = append(searchCriteriaNode, laser.List{laser.Atom("maxPushDate"), laser.Atom(db.searchCriteria.MaxPushDate.Format(time.DateOnly))})
	searchCriteriaNode = append(searchCriteriaNode, laser.List{laser.Atom("maxStars"), laser.Atom(strconv.Itoa(db.searchCriteria.MaxStars))})

	err = searchCriteriaNode.Serialize(output)
	if err != nil {
		return err
	}
	_, err = output.WriteString("\n")
	if err != nil {
		return err
	}

	if db.full {
		_, err = output.WriteString("(full)\n")
		if err != nil {
			return err
		}
	}

	for _, repo := range db.repositories {
		err = laser.List{laser.Atom("repository"), laser.Atom(repo.Id), laser.Atom(repo.CommitHash)}.Serialize(output)
		if err != nil {
			return err
		}
		_, err = output.WriteString("\n")
		if err != nil {
			return err
		}
	}

	return nil
}
