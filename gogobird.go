package main

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"reflect"

	"github.com/ChimeraCoder/anaconda"
	"github.com/mitchellh/cli"
)

var (
	api      *anaconda.TwitterApi
	ui       *cli.ColoredUi
	userInfo UserInfo
	config   *Config
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

func testCredentials() error {
	config = getConfig()
	if config.ConsumerKey == "" || config.ConsumerSecret == "" {
		return errors.New("Credentials are invalid: at least one is empty")
	}
	return nil
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

func newTwitterApi(token string, secret string) *anaconda.TwitterApi {
	api = anaconda.NewTwitterApi(token, secret)
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

func initTwitterApi() bool {
	userInfo, err := ReadAccessCredential()
	if err != nil {
		fmt.Printf("read Access Credentail fail: %v\n", err)
		return false
	}

	api = newTwitterApi(userInfo.Token, userInfo.Secret)
	if api == nil {
		fmt.Printf("Twitter Api client has empty (nil) credentials")
		return false
	}

	ui.Info(fmt.Sprintf("Init Twitter API for [%s] success!\n", userInfo.Name))
	return true
}

/***********************************************************************/
type cmdSearch int

func (cmd cmdSearch) Help() string {
	return "search help"
}

func (cmd cmdSearch) Synopsis() string {
	return "Used to search topic in twitter"
}

func (cmd cmdSearch) Run(args []string) int {
	str, err := ui.Ask("Please enter search string:")
	if err == nil {
		doSearch(api, str)
	}
	return 0
}

func factorySearch() (cli.Command, error) {
	var cmd cmdSearch
	return cmd, nil
}

/***********************************************************************/
type cmdAuth int

func (cmd cmdAuth) Help() string {
	return "authenticate using PIN-based method"
}

func (cmd cmdAuth) Synopsis() string {
	return cmd.Help()
}

func (cmd cmdAuth) Run(args []string) int {
	ui.Output("Authentication start ...")

	url, err := GetAuthUrl()
	if err != nil {
		ui.Error(fmt.Sprintf("Get Authorization URL fail: %v\n", err))
		return -1
	}

	ui.Info(fmt.Sprintf("Please open this URL: %s, and get the PIN code.", url))
	verifier, err := ui.Ask("Please input the PIN code:")
	if err != nil {
		ui.Error("PIN code input invalid")
		return -1
	}

	name, ret := DoAuth(verifier)
	if ret == false {
		ui.Error("DoAuth failed.")
		return -2
	}

	ui.Info(fmt.Sprintf("authenticate [%s] success!\n", name))
	return 0
}

func factoryAuth() (cli.Command, error) {
	var cmd cmdAuth
	return cmd, nil
}

/***********************************************************************/
type cmdPost int

func (cmd cmdPost) Help() string {
	return "post a tweet"
}

func (cmd cmdPost) Synopsis() string {
	return cmd.Help()
}

func (cmd cmdPost) Run(args []string) int {
	tweet, err := ui.Ask("Please input the tweet:\n")
	if err != nil {
		ui.Error("PIN code input invalid")
		return -1
	}

	if postTweet(tweet) == false {
		ui.Error("post tweet failed.")
		return -2
	}
	ui.Info("post tweet success!")
	return 0
}

func factoryPost() (cli.Command, error) {
	var cmd cmdPost
	return cmd, nil
}

/***********************************************************************/
type cmdGetFollowers int

func (cmd cmdGetFollowers) Help() string {
	return "get followers [user_name]"
}

func (cmd cmdGetFollowers) Synopsis() string {
	return cmd.Help()
}

func factoryGetFollowers() (cli.Command, error) {
	var cmd cmdGetFollowers
	return cmd, nil
}

func (cmd cmdGetFollowers) Run(args []string) int {
	vals := url.Values{}

	// to get args[0]'s followers list
	if len(args) == 1 {
		vals.Set("screen_name", args[0])
	}

	ui.Info("Followers:")

	for {
		usersCursor, err := api.GetFollowersList(vals)
		if err != nil {
			ui.Error(fmt.Sprintf("get followers fail: %v", err))
			return -1
		}

		for _, u := range usersCursor.Users {
			ui.Info(fmt.Sprintf("    @%s, %s", u.ScreenName, u.Name))
		}

		if usersCursor.Next_cursor != 0 {
			yn, _ := ui.Ask("continue: y/n ? : ")
			if yn == "n" {
				break
			}
		} else {
			break
		}
		vals.Set("cursor", usersCursor.Next_cursor_str)
	}
	return 0
}

////////////////////////////////////////////////////////////////////////////////

func initUi() {
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
}

func main() {
	if err := testCredentials(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	anaconda.SetConsumerKey(config.ConsumerKey)
	anaconda.SetConsumerSecret(config.ConsumerSecret)
	initUi()

	c := cli.NewCLI("gogobird", "0.0.1")
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		"auth":      factoryAuth,
		"search":    factorySearch,
		"post":      factoryPost,
		"followers": factoryGetFollowers,
	}

	if len(c.Args) >= 1 && os.Args[1] != "auth" {
		if initTwitterApi() == false {
			ui.Error("init twitter api failed.")
			return
		}
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	os.Exit(exitStatus)
}
