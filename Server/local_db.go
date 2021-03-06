package server

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net"
	"os"
	"sync"
)

// Local DB Filename
const (
	constLocalDbFn = "local_db.json"
)

// LocalDb - Local Database
type LocalDb struct {
	users map[string]*UserInfo
	mutex sync.RWMutex
}

// UserInfo - User Info
type UserInfo struct {
	// User nickname
	Name string
	// Password
	Md5Password string

	// Connection to send messages from other users
	// (conn==nil before login and after logouy)
	conn net.Conn

	// It prevents multiple clients from writing messages to the same recipient simultaneously
	mutex sync.Mutex
}

// Init - Initiate Local Db
func (db *LocalDb) Init() error {

	db.users = make(map[string]*UserInfo)

	// create db file if not exist
	if err := db.createIfNotExist(); err != nil {
		return err
	}

	// load from file
	if err := db.load(); err != nil {
		return err
	}

	return nil
}

// createIfNotExist - create db file if not exist
func (db *LocalDb) createIfNotExist() error {

	if _, err := os.Stat(constLocalDbFn); os.IsNotExist(err) || os.IsNotExist(err) {

		// encode json
		data, _ := json.MarshalIndent(db.users, "", " ")

		// write file
		if err := ioutil.WriteFile(constLocalDbFn, data, 0660); err != nil {
			log.Fatal(err)
			return err
		}
	}

	return nil
}

// load - load db fromfile
func (db *LocalDb) load() error {

	db.mutex.Lock()
	defer db.mutex.Unlock()

	// read file
	data, err := ioutil.ReadFile(constLocalDbFn)
	if err != nil {
		log.Fatal(err)
		return err
	}

	// decode json
	db.users = make(map[string]*UserInfo)
	if err := json.Unmarshal(data, &db.users); err != nil {
		return err
	}

	return nil
}

// save to file
func (db *LocalDb) save() error {

	if Debug {
		log.Printf("db: %+v\n", db)
	}

	// encode json
	data, _ := json.MarshalIndent(db.users, "", " ")
	if Debug {
		log.Println(string(data))
	}

	// write file
	if err := ioutil.WriteFile(constLocalDbFn, data, 0660); err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

// RLock - lock for reading
func (db *LocalDb) RLock() {
	db.mutex.RLock()
}

// RUnlock - unlock for reading
func (db *LocalDb) RUnlock() {
	db.mutex.RUnlock()
}

// FindUser - find user
func (db *LocalDb) FindUser(name string) (*UserInfo, bool) {
	val, ok := db.users[name]
	return val, ok
}

// DoesUserExist - check that user exists
func (db *LocalDb) DoesUserExist(name string) bool {
	_, ok := db.users[name]
	return ok
}

// AddUser - Add User
func (db *LocalDb) AddUser(name, password string, conn net.Conn) error {

	db.mutex.Lock()
	defer db.mutex.Unlock()

	// check if user exists
	if db.DoesUserExist(name) {
		return errors.New("User '" + name + "' already exists")
	}

	// add user info
	db.users[name] = &UserInfo{name, password, nil, sync.Mutex{}}

	// save changes
	if err := db.save(); err != nil {
		return err
	}

	return nil
}

// Login - login
func (db *LocalDb) Login(name, password string, conn net.Conn) error {

	db.mutex.Lock()
	defer db.mutex.Unlock()

	user, ok := db.users[name]

	// check if user exists
	if !ok {
		return errors.New("User '" + name + "' does not exist")
	}

	// check if user offline
	if user.conn != nil {
		return errors.New("User '" + name + "' is already online")
	}

	// check password
	if user.Md5Password != password {
		return errors.New("Invalid password")
	}

	// go online
	user.conn = conn
	db.users[name] = user

	return nil
}

// ChangePassword -
func (db *LocalDb) ChangePassword(name, newPassword string) error {

	db.mutex.Lock()
	defer db.mutex.Unlock()

	user, ok := db.users[name]

	// check if user exists
	if !ok {
		return errors.New("User '" + name + "' does not exist")
	}

	// change password
	user.Md5Password = newPassword
	db.users[name] = user
	db.save()

	return nil
}

// GetOnlineUserList - Get Online User List
func (db *LocalDb) GetOnlineUserList() []string {

	db.mutex.RLock()
	defer db.mutex.RUnlock()

	userList := []string{}

	for _, u := range db.users {
		if u.conn != nil {
			userList = append(userList, u.Name)
		}
	}

	return userList
}

// Logout - logout
func (db *LocalDb) Logout(name string) {

	db.mutex.Lock()
	defer db.mutex.Unlock()

	user, ok := db.users[name]
	if ok {
		// go offline
		user.conn = nil
		db.users[name] = user
	}
}

// Clear - Clear Local Db (for testing)
func (db *LocalDb) Clear() {

	db.mutex.Lock()
	defer db.mutex.Unlock()

	// clear db
	db.users = make(map[string]*UserInfo)
}
