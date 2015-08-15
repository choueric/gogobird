package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"reflect"

	"github.com/ChimeraCoder/anaconda"
	"github.com/mitchellh/cli"
)

var (
	CONSUMER_KEY        = os.Getenv("CONSUMER_KEY")
	CONSUMER_SECRET     = os.Getenv("CONSUMER_SECRET")
	ACCESS_TOKEN        = os.Getenv("ACCESS_TOKEN")
	ACCESS_TOKEN_SECRET = os.Getenv("ACCESS_TOKEN_SECRET")
)

var (
	api *anaconda.TwitterApi
	ui  *cli.ColoredUi
)

type User anaconda.User

func (user User) String() string {
	return fmt.Sprintf(" ContributorsEnabled: %v. CreatedAt: %v.\n DefaultProfile: %v. "+
		" DefaultProfileImage: %v.\n Description: %v\n Entities: %v\n"+
		" FavouritesCount: %v. FollowRequestSent: %v. FollowersCount: %v."+
		" Following: %v. FriendsCount: %v.\n GeoEnabled: %v. Id: %v. IdStr: %v.\n"+
		" IsTranslator: %v. Lang: %v. ListedCount: %v. Location: %v. Name: %v.\n"+
		" Notifications: %v. ProfileBackgroundColor: %v.\n ProfileBackgroundImageURL: %v\n"+
		" ProfileBackgroundImageUrlHttps: %v\n ProfileBackgroundTile: %v\n"+
		" ProfileBannerURL: %v\n ProfileImageURL: %v\n ProfileImageUrlHttps: %v\n"+
		" ProfileLinkColor: %v. ProfileSidebarBorderColor: %v.\n"+
		" ProfileSidebarFillColor: %v. ProfileTextColor: %v."+
		" ProfileUseBackgroundImage: %v\n Protected: %v. ScreenName: %v\n"+
		" ShowAllInlineMedia: %v.\n Status: %v\n StatusesCount: %v."+
		" TimeZone: %v. URL: %v\n UtcOffset: %v. Verified: %v."+
		" WithheldInCountries: %v. WithheldScope: %v\n",
		user.ContributorsEnabled, user.CreatedAt, user.DefaultProfile, user.DefaultProfileImage, user.Description,
		user.Entities, user.FavouritesCount, user.FollowRequestSent, user.FollowersCount,
		user.Following, user.FriendsCount, user.GeoEnabled, user.Id, user.IdStr,
		user.IsTranslator, user.Lang, user.ListedCount, user.Location, user.Name,
		user.Notifications, user.ProfileBackgroundColor, user.ProfileBackgroundImageURL,
		user.ProfileBackgroundImageUrlHttps, user.ProfileBackgroundTile, user.ProfileBannerURL,
		user.ProfileImageURL, user.ProfileImageUrlHttps, user.ProfileLinkColor,
		user.ProfileSidebarBorderColor, user.ProfileSidebarFillColor, user.ProfileTextColor,
		user.ProfileUseBackgroundImage, user.Protected, user.ScreenName, user.ShowAllInlineMedia,
		user.Status, user.StatusesCount, user.TimeZone, user.URL, user.UtcOffset, user.Verified,
		user.WithheldInCountries, user.WithheldScope)
}

func postTweet(post string) bool {
	_, err := api.PostTweet(post, url.Values{})
	if err != nil {
		fmt.Printf("post tweet failed: %v\n", err)
		return false
	}
	return true
}

// Test that a valid user can be fetched
// and that unmarshalling works properly
func getUser(username string) bool {
	users, err := api.GetUsersLookup(username, nil)
	if err != nil {
		fmt.Printf("GetUsersLookup returned error: %s", err.Error())
		return false
	}

	if len(users) != 1 {
		fmt.Printf("Expected one user and received %d", len(users))
	}

	// If all attributes are equal to the zero value for that type,
	// then the original value was not valid
	if reflect.DeepEqual(users[0], anaconda.User{}) {
		fmt.Printf("invalid user received\n")
		return false
	}

	for i, user := range users {
		fmt.Printf("---- [%d] ----\n%v\n", i, User(user))
	}
	return true
}

// Test_TwitterCredentials tests that non-empty Twitter credentials are set
// Without this, all following tests will fail
func testCredentials() bool {
	if CONSUMER_KEY == "" || CONSUMER_SECRET == "" || ACCESS_TOKEN == "" || ACCESS_TOKEN_SECRET == "" {
		fmt.Printf("Credentials are invalid: at least one is empty")
		return false
	}
	return true
}

// Test that creating a TwitterApi client creates a client with non-empty OAuth credentials
func newTwitterApi() *anaconda.TwitterApi {
	anaconda.SetConsumerKey(CONSUMER_KEY)
	anaconda.SetConsumerSecret(CONSUMER_SECRET)
	api = anaconda.NewTwitterApi(ACCESS_TOKEN, ACCESS_TOKEN_SECRET)

	if api.Credentials == nil {
		fmt.Printf("Twitter Api client has empty (nil) credentials")
		return nil
	}
	return api
}

func doSearch(api *anaconda.TwitterApi, topic string) {
	// Test that the GetSearch function actually works and returns non-empty results
	search_result, err := api.GetSearch(topic, nil)
	if err != nil {
		ui.Error(fmt.Sprintf("GetSearch yielded error %v", err))
		panic(err)
	}

	// Unless something is seriously wrong, there should be at least two tweets
	if len(search_result.Statuses) < 2 {
		ui.Error(fmt.Sprintf("Expected 2 or more tweets, and found %d", len(search_result.Statuses)))
	}

	// Check that at least one tweet is non-empty
	for i, tweet := range search_result.Statuses {
		if tweet.Text != "" {
			ui.Info(fmt.Sprintf("[%d] %s", i, tweet.Text))
		}
	}
}

func initTwitterApi() {
	if testCredentials() == false {
		return
	}

	if api = newTwitterApi(); api == nil {
		fmt.Printf("Twitter Api client has empty (nil) credentials")
		return
	}

	// doSearch(api, "天津")
	// getUser("chou_eric")

	/*
		if postTweet("test from gogobird") == true {
			fmt.Printf("post tweet success.\n")
		} else {
			fmt.Printf("post tweet fail.\n")
		}
	*/
}

type cmdSearch int

func (cmd cmdSearch) Help() string {
	return "search help"
}

func (cmd cmdSearch) Run(args []string) int {
	str, err := ui.Ask("Please enter search string:")
	if err == nil {
		doSearch(api, str)
	}
	return 0
}

func (cmd cmdSearch) Synopsis() string {
	return "Used to search topic in twitter"
}

func factorySearch() (cli.Command, error) {
	var cmd cmdSearch
	return cmd, nil
}

func main() {
	ui = new(cli.ColoredUi)
	if ui == nil {
		fmt.Printf("error of ui\n")
		return
	}

	bui := new(cli.BasicUi)
	bui.Reader = os.Stdin
	bui.Writer = os.Stdout
	bui.ErrorWriter = os.Stderr

	ui.Ui = bui
	ui.OutputColor = cli.UiColorNone
	ui.InfoColor = cli.UiColorGreen
	ui.ErrorColor = cli.UiColorRed
	ui.WarnColor = cli.UiColorYellow

	c := cli.NewCLI("app", "1.0.0")
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"search": factorySearch,
	}

	initTwitterApi()

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}
