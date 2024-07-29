package database

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type DB struct {
	path string
	mu   *sync.RWMutex
}

type DBStructure struct {
	Chirps        map[int]Chirp           `json:"chirps"`
	Users         map[int]User            `json:"users"`
	RefreshTokens map[string]RefreshToken `json:"refresh_tokens"`
}

type RefreshToken struct {
	ExpiresAt time.Time `json:"expires_at"`
	UserId    int       `json:"user_id"`
}

type Chirp struct {
	ID       int    `json:"id"`
	Body     string `json:"body"`
	AuthorId int    `json:"author_id"`
}

type User struct {
	ID          int    `json:"id"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	IsChirpyRed bool   `json:"is_chirpy_red"`
}

func NewDB(path string) (*DB, error) {
	db := &DB{
		path: path,
		mu:   &sync.RWMutex{},
	}
	err := db.ensureDB()
	return db, err
}

func RemoveDB(path string) error {
	err := os.Remove(path)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) CreateChirp(body, userId string) (Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	id := len(dbStructure.Chirps) + 1
	userIdAsInt, _ := strconv.Atoi(userId)
	chirp := Chirp{
		ID:       id,
		Body:     body,
		AuthorId: userIdAsInt,
	}
	dbStructure.Chirps[id] = chirp

	err = db.writeDB(dbStructure)
	if err != nil {
		return Chirp{}, err
	}

	return chirp, nil
}

func (db *DB) DeleteChirp(chirpId int, userId string) error {
	dbStructure, err := db.loadDB()
	if err != nil {
		return err
	}

	userIdAsInt, _ := strconv.Atoi(userId)

	if userIdAsInt != dbStructure.Chirps[chirpId].AuthorId {
		return errors.New("you are not authorized to delete that chirp")
	}

	dbStructure.Chirps[chirpId] = Chirp{}

	err = db.writeDB(dbStructure)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) GetChirps() ([]Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return nil, err
	}

	chirps := make([]Chirp, 0, len(dbStructure.Chirps))
	for _, chirp := range dbStructure.Chirps {
		chirps = append(chirps, chirp)
	}

	return chirps, nil
}

func (db *DB) GetChirp(id int) (Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	if id > len(dbStructure.Chirps) {
		return Chirp{}, errors.New("Chirp does not exist")
	}

	chirp := dbStructure.Chirps[id]

	return chirp, nil
}

func (db *DB) CreateUser(email, password string) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	id := len(dbStructure.Users) + 1
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 1)
	if err != nil {
		return User{}, err
	}
	user := User{
		ID:       id,
		Email:    email,
		Password: string(hashedPassword),
	}
	dbStructure.Users[id] = user

	err = db.writeDB(dbStructure)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (db *DB) UpdateUser(email, password, userId string) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	id, err := strconv.Atoi(userId)
	if err != nil {
		return User{}, err
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 1)
	if err != nil {
		return User{}, err
	}
	user := User{
		ID:       id,
		Email:    email,
		Password: string(hashedPassword),
	}
	dbStructure.Users[id] = user

	err = db.writeDB(dbStructure)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (db *DB) UpgradeUser(userId int) error {
	dbStructure, err := db.loadDB()
	if err != nil {
		return err
	}

	user, ok := dbStructure.Users[userId]
	if !ok {
		return errors.New("User not found")
	}

	user.IsChirpyRed = true
    dbStructure.Users[userId] = user

	err = db.writeDB(dbStructure)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) AuthenticateUser(email, password string) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	user, err := getUserByEmail(email, &dbStructure)
	if err != nil {
		return User{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (db *DB) GenerateRefreshToken(userId int) (string, error) {
	refreshToken := make([]byte, 32)
	_, _ = rand.Read(refreshToken)
	refreshTokenString := hex.EncodeToString(refreshToken)

	dbStructure, err := db.loadDB()
	if err != nil {
		return "", err
	}

	dbStructure.RefreshTokens[refreshTokenString] = RefreshToken{
		ExpiresAt: time.Now().Add(time.Hour * time.Duration(24) * time.Duration(60)),
		UserId:    userId,
	}

	err = db.writeDB(dbStructure)
	if err != nil {
		return "", err
	}

	_, ok := dbStructure.RefreshTokens[refreshTokenString]
	if !ok {
		return "", errors.New("refresh token was not saved to db")
	}

	return refreshTokenString, nil
}

func (db *DB) ValidateRefreshToken(refreshTokenString string) (int, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return 0, err
	}

	refreshToken, ok := dbStructure.RefreshTokens[refreshTokenString]
	if !ok {
		return 0, errors.New("token does not exist")
	}

	if refreshToken.ExpiresAt.Before(time.Now()) {
		return 0, errors.New("token has expired")
	}

	return refreshToken.UserId, nil
}

func (db *DB) RevokeRefreshToken(refreshTokenString string) error {
	dbStructure, err := db.loadDB()
	if err != nil {
		return err
	}

	_, ok := dbStructure.RefreshTokens[refreshTokenString]
	if !ok {
		return errors.New("token doese not exist")
	}

	delete(dbStructure.RefreshTokens, refreshTokenString)
	err = db.writeDB(dbStructure)
	if err != nil {
		return err
	}
	return nil
}

func getUserByEmail(email string, dbStructure *DBStructure) (User, error) {
	for _, user := range dbStructure.Users {
		if user.Email == email {
			return user, nil
		}
	}
	return User{}, errors.New("couldnt find user by that email")
}

func (db *DB) createDB() error {
	dbStructure := DBStructure{
		Chirps:        map[int]Chirp{},
		Users:         map[int]User{},
		RefreshTokens: map[string]RefreshToken{},
	}
	return db.writeDB(dbStructure)
}

func (db *DB) ensureDB() error {
	_, err := os.ReadFile(db.path)
	if errors.Is(err, os.ErrNotExist) {
		return db.createDB()
	}
	return err
}

func (db *DB) loadDB() (DBStructure, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	dbStructure := DBStructure{}
	dat, err := os.ReadFile(db.path)
	if errors.Is(err, os.ErrNotExist) {
		return dbStructure, err
	}
	err = json.Unmarshal(dat, &dbStructure)
	if err != nil {
		return dbStructure, err
	}

	return dbStructure, nil
}

func (db *DB) writeDB(dbStructure DBStructure) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	dat, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}

	err = os.WriteFile(db.path, dat, 0600)
	if err != nil {
		return err
	}
	return nil
}
