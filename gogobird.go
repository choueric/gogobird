package main

import (
	"fmt"
	"net/url"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/ChimeraCoder/anaconda"
)

type User anaconda.User

var CONSUMER_KEY = os.Getenv("CONSUMER_KEY")
var CONSUMER_SECRET = os.Getenv("CONSUMER_SECRET")
var ACCESS_TOKEN = os.Getenv("ACCESS_TOKEN")
var ACCESS_TOKEN_SECRET = os.Getenv("ACCESS_TOKEN_SECRET")

var api *anaconda.TwitterApi

func Test_GetFavorites(t *testing.T) {
	v := url.Values{}
	v.Set("screen_name", "chimeracoder")
	favorites, err := api.GetFavorites(v)
	if err != nil {
		t.Errorf("GetFavorites returned error: %s", err.Error())
	}

	if len(favorites) == 0 {
		t.Errorf("GetFavorites returned no favorites")
	}

	if reflect.DeepEqual(favorites[0], anaconda.Tweet{}) {
		t.Errorf("GetFavorites returned %d favorites and the first one was empty", len(favorites))
	}
}

// Test that a valid tweet can be fetched properly
// and that unmarshalling of tweet works without error
func Test_GetTweet(t *testing.T) {
	const tweetId = 303777106620452864
	const tweetText = `golang-syd is in session. Dave Symonds is now talking about API design and protobufs. #golang http://t.co/eSq3ROwu`

	tweet, err := api.GetTweet(tweetId, nil)
	if err != nil {
		t.Errorf("GetTweet returned error: %s", err.Error())
	}

	if tweet.Text != tweetText {
		t.Errorf("Tweet %d contained incorrect text. Received: %s", tweetId, tweetText)
	}

	// Check the entities
	expectedEntities := anaconda.Entities{Hashtags: []struct {
		Indices []int
		Text    string
	}{struct {
		Indices []int
		Text    string
	}{Indices: []int{86, 93}, Text: "golang"}}, Urls: []struct {
		Indices      []int
		Url          string
		Display_url  string
		Expanded_url string
	}{}, User_mentions: []struct {
		Name        string
		Indices     []int
		Screen_name string
		Id          int64
		Id_str      string
	}{}, Media: []anaconda.EntityMedia{anaconda.EntityMedia{Id: 303777106628841472, Id_str: "303777106628841472", Media_url: "http://pbs.twimg.com/media/BDc7q0OCEAAoe2C.jpg", Media_url_https: "https://pbs.twimg.com/media/BDc7q0OCEAAoe2C.jpg", Url: "http://t.co/eSq3ROwu", Display_url: "pic.twitter.com/eSq3ROwu", Expanded_url: "http://twitter.com/golang/status/303777106620452864/photo/1", Sizes: anaconda.MediaSizes{Medium: anaconda.MediaSize{W: 600, H: 450, Resize: "fit"}, Thumb: anaconda.MediaSize{W: 150, H: 150, Resize: "crop"}, Small: anaconda.MediaSize{W: 340, H: 255, Resize: "fit"}, Large: anaconda.MediaSize{W: 1024, H: 768, Resize: "fit"}}, Type: "photo", Indices: []int{94, 114}}}}
	if !reflect.DeepEqual(tweet.Entities, expectedEntities) {
		t.Errorf("Tweet entities differ")
	}

}

// This assumes that the current user has at least two pages' worth of followers
func Test_GetFollowersListAll(t *testing.T) {
	result := api.GetFollowersListAll(nil)
	i := 0

	for page := range result {
		if i == 2 {
			return
		}

		if page.Error != nil {
			t.Errorf("Receved error from GetFollowersListAll: %s", page.Error)
		}

		if page.Followers == nil || len(page.Followers) == 0 {
			t.Errorf("Received invalid value for page %d of followers: %v", i, page.Followers)
		}
		i++
	}
}

// This assumes that the current user has at least two pages' worth of friends
func Test_GetFriendsIdsAll(t *testing.T) {
	result := api.GetFriendsIdsAll(nil)
	i := 0

	for page := range result {
		if i == 2 {
			return
		}

		if page.Error != nil {
			t.Errorf("Receved error from GetFriendsIdsAll : %s", page.Error)
		}

		if page.Ids == nil || len(page.Ids) == 0 {
			t.Errorf("Received invalid value for page %d of friends : %v", i, page.Ids)
		}
		i++
	}
}

// Test that setting the delay actually changes the stored delay value
func Test_TwitterApi_SetDelay(t *testing.T) {
	const OLD_DELAY = 1 * time.Second
	const NEW_DELAY = 20 * time.Second
	api.EnableThrottling(OLD_DELAY, 4)

	delay := api.GetDelay()
	if delay != OLD_DELAY {
		t.Errorf("Expected initial delay to be the default delay (%s)", anaconda.DEFAULT_DELAY.String())
	}

	api.SetDelay(NEW_DELAY)

	if newDelay := api.GetDelay(); newDelay != NEW_DELAY {
		t.Errorf("Attempted to set delay to %s, but delay is now %s (original delay: %s)", NEW_DELAY, newDelay, delay)
	}
}

func Test_TwitterApi_TwitterErrorDoesNotExist(t *testing.T) {

	// Try fetching a tweet that no longer exists (was deleted)
	const DELETED_TWEET_ID = 404409873170841600

	tweet, err := api.GetTweet(DELETED_TWEET_ID, nil)
	if err == nil {
		t.Errorf("Expected an error when fetching tweet with id %d but got none - tweet object is %+v", DELETED_TWEET_ID, tweet)
	}

	apiErr, ok := err.(*anaconda.ApiError)
	if !ok {
		t.Errorf("Expected an *anaconda.ApiError, and received error message %s, (%+v)", err.Error(), err)
	}

	terr, ok := apiErr.Decoded.First().(anaconda.TwitterError)

	if !ok {
		t.Errorf("TwitterErrorResponse.First() should return value of type TwitterError, not %s", reflect.TypeOf(apiErr.Decoded.First()))
	}

	if code := terr.Code; code != anaconda.TwitterErrorDoesNotExist && code != anaconda.TwitterErrorDoesNotExist2 {
		if code == anaconda.TwitterErrorRateLimitExceeded {
			t.Errorf("Rate limit exceeded during testing - received error code %d instead of %d", anaconda.TwitterErrorRateLimitExceeded, anaconda.TwitterErrorDoesNotExist)
		} else {

			t.Errorf("Expected Twitter to return error code %d, and instead received error code %d", anaconda.TwitterErrorDoesNotExist, code)
		}
	}
}

// Test that the client can be used to throttle to an arbitrary duration
func Test_TwitterApi_Throttling(t *testing.T) {
	const MIN_DELAY = 15 * time.Second

	api.EnableThrottling(MIN_DELAY, 5)
	oldDelay := api.GetDelay()
	api.SetDelay(MIN_DELAY)

	now := time.Now()
	_, err := api.GetSearch("golang", nil)
	if err != nil {
		t.Errorf("GetSearch yielded error %s", err.Error())
	}
	_, err = api.GetSearch("anaconda", nil)
	if err != nil {
		t.Errorf("GetSearch yielded error %s", err.Error())
	}
	after := time.Now()

	if difference := after.Sub(now); difference < MIN_DELAY {
		t.Errorf("Expected delay of at least %s. Actual delay: %s", MIN_DELAY.String(), difference.String())
	}

	// Reset the delay to its previous value
	api.SetDelay(oldDelay)
}

/***************************************************************************/
func getDMScreenName() {
	to, err := api.GetSelf(url.Values{})
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
	_, err = api.PostDMToScreenName("Test the anaconda lib", to.ScreenName)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}
}

func postTweet(post string) bool {
	_, err := api.PostTweet(post, url.Values{})
	if err != nil {
		fmt.Printf("post tweet failed: %v\n", err)
		return false
	}
	return true
}

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

func trySearch(api *anaconda.TwitterApi, topic string) {
	// Test that the GetSearch function actually works and returns non-empty results
	search_result, err := api.GetSearch(topic, nil)
	if err != nil {
		fmt.Printf("GetSearch yielded error %s", err.Error())
		panic(err)
	}

	// Unless something is seriously wrong, there should be at least two tweets
	if len(search_result.Statuses) < 2 {
		fmt.Printf("Expected 2 or more tweets, and found %d", len(search_result.Statuses))
	}

	// Check that at least one tweet is non-empty
	for i, tweet := range search_result.Statuses {
		if tweet.Text != "" {
			fmt.Println(i, tweet.Text)
		}
	}
}

func main() {
	if testCredentials() == false {
		return
	}

	if api = newTwitterApi(); api == nil {
		fmt.Printf("Twitter Api client has empty (nil) credentials")
		return
	}

	//trySearch(api, "天津")
	//getUser("chou_eric")
	if postTweet("test from gogobird") == true {
		fmt.Printf("post tweet success.\n")
	} else {
		fmt.Printf("post tweet fail.\n")
	}
}
