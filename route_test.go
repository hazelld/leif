package leif_test

import (
	"fmt"
	"io/ioutil"
	. "leif"
	"mma/api"
	"mma/middleware"
	"os"
	"testing"
)

func TestRoute(t *testing.T) {
	RegisterHandler(api.GetTeam,
		api.GetTeamDraft,
		api.GetTeamPoints,
		api.GetTeamRoster,
		api.GetTeamTrade,
		api.GetTeamTransactions, api.GetFighters, api.GetFightersById)

	RegisterMiddleware(middleware.ValidateAPIMiddleware,
		middleware.LoadUserMiddleware, middleware.ValidateHTMLRequestMiddleware)
	file, _ := os.Open("test.json")
	defer file.Close()
	byteJson, _ := ioutil.ReadAll(file)
	routes, err := Parse(byteJson)

	if err != nil {
		t.Errorf(err.Error())
	}

	for _, r := range routes {
		fmt.Println(r)
	}
}
