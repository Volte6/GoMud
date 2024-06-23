package users

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"log/slog"

	"github.com/volte6/mud/configs"
	"github.com/volte6/mud/connection"
	"github.com/volte6/mud/mobs"
	"github.com/volte6/mud/util"

	//
	"gopkg.in/yaml.v2"
)

const minimumUsernameLength = 2
const maximumUsernameLength = 16
const minimumPasswordLength = 4
const maximumPasswordLength = 16

var (
	userManager *ActiveUsers = NewUserManager()
)

type ActiveUsers struct {
	sync.RWMutex
	Users             map[int]*UserRecord                // userId to UserRecord
	Usernames         map[string]int                     // username to userId
	Connections       map[connection.ConnectionId]int    // connectionId to userId
	UserConnections   map[int]connection.ConnectionId    // userId to connectionId
	ZombieConnections map[connection.ConnectionId]uint64 // connectionId to turn they became a zombie
}

type Online struct {
	UserId         string
	Username       string
	Permission     string
	CharacterName  string
	CharacterLevel int
	Zone           string
	RoomId         int
}

func NewUserManager() *ActiveUsers {
	return &ActiveUsers{
		Users:             make(map[int]*UserRecord),
		Usernames:         make(map[string]int),
		Connections:       make(map[connection.ConnectionId]int),
		UserConnections:   make(map[int]connection.ConnectionId),
		ZombieConnections: make(map[connection.ConnectionId]uint64),
	}
}

func RemoveZombieUser(userId int) {
	userManager.Lock()
	defer userManager.Unlock()

	connId := userManager.UserConnections[userId]
	delete(userManager.ZombieConnections, connId)
}

func RemoveZombieConnection(connectionId connection.ConnectionId) {
	userManager.Lock()
	defer userManager.Unlock()

	delete(userManager.ZombieConnections, connectionId)
}

func GetExpiredZombies(expirationTurn uint64) []int {
	userManager.Lock()
	defer userManager.Unlock()

	expiredUsers := make([]int, 0)

	for connectionId, zombieTurn := range userManager.ZombieConnections {
		if zombieTurn < expirationTurn {
			slog.Info("TEST", "expiration", expirationTurn, "zombieTurn", zombieTurn)
			expiredUsers = append(expiredUsers, userManager.Connections[connectionId])
		}
	}

	return expiredUsers
}

func GetConnectionIds(userIds []int) []connection.ConnectionId {
	userManager.RLock()
	defer userManager.RUnlock()

	connectionIds := make([]connection.ConnectionId, 0, len(userIds))
	for _, userId := range userIds {
		if user, ok := userManager.Users[userId]; ok {
			connectionIds = append(connectionIds, user.connectionId)
		}
	}

	return connectionIds
}

func GetOnlineUserIds() []int {
	userManager.RLock()
	defer userManager.RUnlock()

	onlineList := make([]int, 0, len(userManager.Users))
	for _, user := range userManager.Users {
		onlineList = append(onlineList, user.UserId)
	}
	return onlineList
}

func GetOnlineList() []Online {

	userManager.RLock()
	defer userManager.RUnlock()

	onlineList := make([]Online, 0, len(userManager.Users))
	for _, user := range userManager.Users {
		onlineList = append(onlineList, Online{
			UserId:         strconv.Itoa(int(user.UserId)),
			Username:       user.Username,
			Permission:     user.Permission,
			CharacterName:  user.Character.Name,
			CharacterLevel: user.Character.Level,
			Zone:           user.Character.Zone,
			RoomId:         user.Character.RoomId,
		})
	}

	return onlineList
}

func GetByCharacterName(name string) *UserRecord {
	userManager.RLock()
	defer userManager.RUnlock()

	var closeMatch *UserRecord = nil

	name = strings.ToLower(name)
	for _, user := range userManager.Users {
		testName := strings.ToLower(user.Character.Name)
		if testName == name {
			return user
		}
		if strings.HasPrefix(testName, name) {
			closeMatch = user
		}
	}

	return closeMatch
}

func GetByUserId(userId int) *UserRecord {
	userManager.RLock()
	defer userManager.RUnlock()

	if user, ok := userManager.Users[userId]; ok {
		return user
	}

	return nil
}

func GetByConnectionId(connectionId connection.ConnectionId) *UserRecord {
	userManager.RLock()
	defer userManager.RUnlock()

	if userId, ok := userManager.Connections[connectionId]; ok {
		return userManager.Users[userId]
	}

	return nil
}

// First time creating a user.
func LoginUser(u *UserRecord, connectionId connection.ConnectionId) (string, error) {

	slog.Info("Logging in user", "username", u.Username, "connectionId", connectionId)

	userManager.Lock()
	defer userManager.Unlock()

	if userId, ok := userManager.Usernames[u.Username]; ok {

		if otherConnId, ok := userManager.UserConnections[userId]; ok {

			if _, ok := userManager.ZombieConnections[otherConnId]; ok {

				// The user is a zombie.
				delete(userManager.ZombieConnections, otherConnId)

				u.connectionId = connectionId

				userManager.Users[u.UserId] = u
				userManager.Usernames[u.Username] = u.UserId
				userManager.Connections[u.connectionId] = u.UserId
				userManager.UserConnections[u.UserId] = u.connectionId

				return "Reconnecting...", nil
			}

		}

		return "That user is already logged in.", errors.New("user is already logged in")
	}

	if len(u.AdminCommands) > 0 {
		u.Permission = PermissionMod
	}

	u.connectionId = connectionId

	userManager.Users[u.UserId] = u
	userManager.Usernames[u.Username] = u.UserId
	userManager.Connections[u.connectionId] = u.UserId
	userManager.UserConnections[u.UserId] = u.connectionId

	return "", nil
}

func SetZombieConnection(connId connection.ConnectionId) {

	userManager.Lock()
	defer userManager.Unlock()

	if _, ok := userManager.ZombieConnections[connId]; ok {
		return
	}

	userManager.ZombieConnections[connId] = util.GetTurnCount()
}

func SetZombieUser(userId int) {

	userManager.Lock()
	defer userManager.Unlock()

	if u, ok := userManager.Users[userId]; ok {

		if _, ok := userManager.ZombieConnections[u.connectionId]; ok {
			return
		}

		userManager.ZombieConnections[u.connectionId] = util.GetTurnCount()
	}

}

func SaveAllUsers() {
	userManager.Lock()
	defer userManager.Unlock()

	for _, u := range userManager.Users {
		if err := SaveUser(*u); err != nil {
			slog.Error("SaveAllUsers()", "error", err.Error())
		}
	}

}

func LogOutUserByConnectionId(connectionId connection.ConnectionId) error {

	u := GetByConnectionId(connectionId)

	userManager.Lock()
	defer userManager.Unlock()

	if _, ok := userManager.Connections[connectionId]; ok {

		// Make sure the user data is saved to a file.
		if u != nil {
			u.Character.Validate()
			SaveUser(*u)
		}

		delete(userManager.Users, u.UserId)
		delete(userManager.Usernames, u.Username)
		delete(userManager.Connections, u.connectionId)
		delete(userManager.UserConnections, u.UserId)

		return nil
	}
	return errors.New("user not found for connection")
}

// First time creating a user.
func CreateUser(u *UserRecord) error {

	if configs.GetConfig().IsBannedName(u.Username) {
		return errors.New("that username is not allowed")
	}

	for _, name := range mobs.GetAllMobNames() {
		if strings.EqualFold(name, u.Username) {
			return errors.New("that username is not allowed")
		}
	}

	if Exists(u.Username) {
		return errors.New("user already exists")
	}

	u.UserId = GetUniqueUserId()
	u.Permission = PermissionUser

	if err := SaveUser(*u); err != nil {
		return err
	}

	userManager.Lock()
	defer userManager.Unlock()

	userManager.Users[u.UserId] = u
	userManager.Usernames[u.Username] = u.UserId
	userManager.Connections[u.connectionId] = u.UserId
	userManager.UserConnections[u.UserId] = u.connectionId

	return nil
}

func LoadUser(username string) (*UserRecord, error) {
	if !Exists(strings.ToLower(username)) {
		return nil, errors.New("user already exists")
	}

	slog.Info("Loading user", "username", username)

	userFilePath := util.FilePath(string(configs.GetConfig().FolderUserData), `/`, strings.ToLower(username)+`.yaml`)

	userFileTxt, err := os.ReadFile(userFilePath)
	if err != nil {
		return nil, err
	}

	loadedUser := &UserRecord{}
	if err := yaml.Unmarshal([]byte(userFileTxt), loadedUser); err != nil {
		slog.Error("LoadUser", "error", err.Error())
	}

	rebuiltMemory := []int{}
	memoryString := string(util.Decompress(util.Decode(loadedUser.RoomMemoryBlob)))
	for _, rId := range strings.Split(memoryString, ",") {
		if rIdInt, err := strconv.Atoi(rId); err == nil {
			rebuiltMemory = append(rebuiltMemory, rIdInt)
		}
	}

	loadedUser.Character.SetRoomMemory(rebuiltMemory)
	loadedUser.RoomMemoryBlob = ``

	if err := loadedUser.Character.Validate(); err == nil {
		SaveUser(*loadedUser)
	}

	// Set their connection time to now
	loadedUser.connectionTime = time.Now()

	return loadedUser, nil
}

func SaveUser(u UserRecord) error {

	slog.Info("Saving user", "username", u.Username)

	memoryString := ``
	for _, rId := range u.Character.GetRoomMemory() {
		memoryString += strconv.Itoa(rId) + ","
	}
	memoryString = strings.TrimSuffix(memoryString, ",")

	u.RoomMemoryBlob = util.Encode(util.Compress([]byte(memoryString)))

	data, err := yaml.Marshal(&u)
	if err != nil {
		return err
	}

	carefulSave := configs.GetConfig().CarefulSaveFiles

	path := util.FilePath(string(configs.GetConfig().FolderUserData), `/`, strings.ToLower(u.Username)+`.yaml`)

	saveFilePath := path
	if carefulSave { // careful save first saves a {filename}.new file
		saveFilePath += `.new`
	}

	err = os.WriteFile(saveFilePath, data, 0777)
	if err != nil {
		return err
	}

	if carefulSave {
		//
		// Once the file is written, rename it to remove the .new suffix and overwrite the old file
		//
		if err := os.Rename(saveFilePath, path); err != nil {
			return err
		}
	}

	slog.Info("Saved user", "username", u.Username)

	return nil
}

func GetUniqueUserId() int {
	return UserCount() + 1
}

func Exists(name string) bool {
	_, err := os.Stat(util.FilePath(string(configs.GetConfig().FolderUserData), `/`, strings.ToLower(name)+`.yaml`))
	return !os.IsNotExist(err)
}

func UserCount() int {

	entries, err := os.ReadDir(util.FilePath(string(configs.GetConfig().FolderUserData)))
	if err != nil {
		panic(err)
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			count++
		}
	}
	return count
}
