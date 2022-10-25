package database

import (
	"errors"
	"fmt"
	"time"
	"yacoid_server/common"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	gomail "gopkg.in/gomail.v2"
)

type User struct {
	ID                               primitive.ObjectID `bson:"_id" json:"-"`
	Admin                            bool               `bson:"admin" json:"-"`
	RegistrationDate                 time.Time          `bson:"registration_date" json:"registrationDate"`
	FirstName                        string             `bson:"first_name" json:"firstName" validate:"required,min=1"`
	LastName                         string             `bson:"last_name" json:"lastName" validate:"required,min=1"`
	Email                            string             `bson:"email" json:"email" validate:"required,email"`
	PasswordHash                     string             `bson:"password_hash" json:"passwordHash" validate:"required"`
	PasswordSalt                     string             `bson:"password_salt" json:"passwordSalt"`
	AuthToken                        string             `bson:"auth_token,omitempty" json:"authToken,omitempty"`
	PasswordResetToken               *string            `bson:"password_reset_token,omitempty" json:"-"`
	PasswordResetTokenExpiryDate     *time.Time         `bson:"password_reset_token_expiry_date,omitempty" json:"-"`
	PendingEmail                     *string            `bson:"pending_email,omitempty" json:"pendingEmail,omitempty"`
	EmailVerificationToken           *string            `bson:"email_verification_token,omitempty" json:"-"`
	EmailVerificationTokenExpiryDate *time.Time         `bson:"email_verification_token_expiry_date,omitempty" json:"-"`
}

func (user *User) Validate(validate *validator.Validate) []string {
	return common.ValidateStruct(user, validate)
}

var ErrorUnknown = errors.New("UNKNOWN_ERROR")
var ErrorInvalidCredentials = errors.New("INVALID_CREDENTIALS")
var ErrorInvalidAuthToken = errors.New("INVALID_AUTH_TOKEN")
var ErrorUserAlreadyExists = errors.New("USER_ALREADY_EXISTS")
var ErrorUserAlreadyLoggedIn = errors.New("USER_ALREADY_LOGGED_IN")
var ErrorUserNotLoggedIn = errors.New("USER_NOT_LOGGED_IN")
var ErrorPasswordResetExpiryDateExceeded = errors.New("PASSWORD_RESET_EXPIRY_DATE_EXCEEDED")

var ErrorEmailVerificationToken = errors.New("EMAIL_VERIFICATION_TOKEN_ERROR")
var ErrorChangePasswordToken = errors.New("CHANGE_PASSWORD_TOKEN_ERROR")

func Login(email string, passwordHash string) (*User, error) {

	fmt.Println(email, passwordHash)
	user, err := GetUserByEmail(email)

	if err != nil {
		return nil, err
	}

	if isCorrectPassword(user, passwordHash) {

		newUser, authError := UpdateUserAuth(user)
		if authError != nil {
			return nil, authError
		}

		return newUser, nil
	} else {
		return nil, ErrorInvalidCredentials
	}

}

func GetPasswordSalt(email string) (*string, error) {

	user, err := GetUserByEmail(email)

	if err != nil {
		if err == ErrorUserNotFound {
			randomID := seededUUID(email) // send a fake uuid to not expose if an user does not exist
			return &randomID, nil
		}
		return nil, err
	}

	return &user.PasswordSalt, nil

}

func Logout(authToken string) error {

	user, findError := GetUserByAuthToken(authToken)
	if findError != nil {
		if findError == mongo.ErrNoDocuments {
			return ErrorUserNotLoggedIn
		}
		return findError
	}

	filter := bson.M{"_id": user.ID}
	update := bson.M{
		"$set": bson.M{"auth_token": nil},
	}

	var currentUser User
	err := userCollection.FindOneAndUpdate(dbContext, filter, update).Decode(&currentUser)

	return err

}

func Register(user User) (*User, error) {

	userExists, findError := UserExists(user.Email)

	if findError != nil {
		fmt.Println(findError.Error())
		return nil, findError
	}

	if userExists {
		return nil, ErrorUserAlreadyExists
	}

	user.ID = primitive.NewObjectID()
	user.Admin = false
	user.RegistrationDate = time.Now()
	user.PasswordSalt = uuid.NewString()
	_, err := userCollection.InsertOne(dbContext, &user)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

func UpdateUserAuth(user *User) (*User, error) {

	/*currentUser, authError := GetUserById(user.ID)

	if authError != nil {

		if authError == mongo.ErrNoDocuments {
			return nil, UnknownError
		}

		return nil, authError
	}

	if currentUser.AuthToken != "" {
		return nil, UserAlreadyLoggedInError
	}*/

	filter := bson.M{"_id": user.ID}

	id := uuid.NewString()

	update := bson.M{
		"$set": bson.M{"auth_token": id},
	}

	after := options.After
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
	}

	var newUser User
	err := userCollection.FindOneAndUpdate(dbContext, filter, update, &opt).Decode(&newUser)

	if err != nil {
		return nil, err
	}

	return &newUser, nil

}

func UpdateExpoPushToken(authToken string, expoPushToken string) error {

	user, userError := GetUserByAuthToken(authToken)

	if userError != nil {
		return userError
	}

	filter := bson.M{"_id": user.ID}

	update := bson.M{
		"$set": bson.M{"expo_push_token": expoPushToken},
	}

	options := options.FindOneAndUpdate()
	options.SetUpsert(true)

	var document bson.D
	err := userCollection.FindOneAndUpdate(dbContext, filter, update, options).Decode(&document)

	return err

}

func InitiatePasswordReset(email string) (*string, error) {

	user, findError := GetUserByEmail(email)

	if findError != nil {
		return nil, findError
	}

	filter := bson.M{"_id": user.ID}

	passwordResetToken := uuid.NewString()

	update := bson.M{
		"$set": bson.M{"password_reset_token": passwordResetToken, "password_reset_token_expiry_date": time.Now().Add(time.Minute * 60 * 24 * 7)},
	}

	after := options.After
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
	}

	var updatedUser User
	err := userCollection.FindOneAndUpdate(dbContext, filter, update, &opt).Decode(&updatedUser)

	if err != nil {
		return nil, err
	}

	sendPasswordResetEmail(email, passwordResetToken)

	return &passwordResetToken, nil

}

func sendPasswordResetEmail(email string, passwordResetToken string) error {
	return sendMail(email, "Password zurücksetzen", "<b>Klicke auf diesen Link, um dein Passwort zurückzusetzen:<b><br/><a href=\"http://localhost:3000/reset_password/"+passwordResetToken+"\">Passwort zurücksetzen</a>")
}

func ResetPassword(passwordResetToken string, passwordHash string) error {

	user, findError := GetUserByPasswordResetToken(passwordResetToken)

	if findError != nil {
		return findError
	}

	if time.Now().After(*user.PasswordResetTokenExpiryDate) {
		return ErrorPasswordResetExpiryDateExceeded
	}

	filter := bson.M{"_id": user.ID}

	update := bson.M{
		"$set": bson.M{"password_reset_token": nil, "password_reset_token_expiry_date": nil, "password_hash": passwordHash},
	}
	fmt.Println("AC", update)

	after := options.After
	opt := options.FindOneAndUpdateOptions{
		ReturnDocument: &after,
	}

	var updatedUser User
	err := userCollection.FindOneAndUpdate(dbContext, filter, update, &opt).Decode(&updatedUser)

	if err != nil {
		return err
	}

	return nil

}

func GetUserById(id primitive.ObjectID) (*User, error) {

	filter := bson.M{"_id": id}
	return GetUserByFilter(filter)

}

func GetUserByEmail(email string) (*User, error) {

	filter := bson.M{"email": email}
	return GetUserByFilter(filter)

}

func GetUserByAuthToken(authToken string) (*User, error) {

	filter := bson.M{"auth_token": authToken}
	user, err := GetUserByFilter(filter)

	if err != nil {
		if err == ErrorUserNotFound {
			return nil, ErrorInvalidAuthToken
		}
		return nil, err
	}
	return user, nil

}

func GetUserByPasswordResetToken(passwordResetToken string) (*User, error) {

	filter := bson.M{"password_reset_token": passwordResetToken}
	return GetUserByFilter(filter)

}

func GetUserByFilter(filter primitive.M) (*User, error) {

	var user User
	err := userCollection.FindOne(dbContext, filter).Decode(&user)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrorUserNotFound
		}
		return nil, err
	}

	return &user, nil

}

func UserExists(email string) (bool, error) {

	user, err := GetUserByEmail(email)

	if err != nil {
		if err == ErrorUserNotFound {
			return false, nil
		}
		return false, err
	}

	return user != nil, err

}

func GetUserCount() (int64, error) {
	return userCollection.CountDocuments(dbContext, bson.M{})
}

func ChangeProjectSaveStatus(authToken string, projectId string, saved bool) error {

	user, err := GetUserByAuthToken(authToken)
	fmt.Println(user, err)

	if err != nil {
		return err
	}

	filter := bson.M{"_id": user.ID}

	projectObjectId, tokenError := primitive.ObjectIDFromHex(projectId)
	fmt.Println(projectObjectId, tokenError)

	if tokenError != nil {
		return tokenError
	}

	var update bson.M

	if saved {
		update = bson.M{
			"$push": bson.M{"favourite_project_ids": projectObjectId},
		}
	} else {
		update = bson.M{
			"$pull": bson.M{"favourite_project_ids": projectObjectId},
		}
	}

	options := options.FindOneAndUpdate()
	options.SetUpsert(true)

	fmt.Println(update)

	var result bson.D
	findError := userCollection.FindOneAndUpdate(dbContext, filter, update, options).Decode(&result)
	fmt.Println(findError)

	return findError

}

type ChangeAccountDataResponse struct {
	FirstName         *UpdateState `bson:"first_name,omitempty" json:"firstName,omitempty"`
	LastName          *UpdateState `bson:"last_name,omitempty" json:"lastName,omitempty"`
	City              *UpdateState `bson:"city,omitempty" json:"city,omitempty"`
	EmailVerification *UpdateState `bson:"email_verification,omitempty" json:"emailVerification,omitempty"`
	ChangePassword    *UpdateState `bson:"change_password,omitempty" json:"changePassword,omitempty"`
}

func ChangeAccountData(authToken string, firstName *string, lastName *string, email *string, city *string, currentPassword *string, newPassword *string) (*ChangeAccountDataResponse, error) {

	user, userError := GetUserByAuthToken(authToken)
	fmt.Println("USER", user, userError)

	if userError != nil {
		return nil, userError
	}

	var response ChangeAccountDataResponse

	options := options.FindOneAndUpdate()
	options.SetUpsert(true)

	filter := bson.M{"_id": user.ID}

	/* Handle simple changes */
	var inputs []UpdateEntry
	inputs = append(inputs, UpdateEntry{field: "first_name", value: firstName})
	inputs = append(inputs, UpdateEntry{field: "last_name", value: lastName})
	inputs = append(inputs, UpdateEntry{field: "city", value: city})
	fmt.Println("INPUTS", inputs)
	updateEntries := CreateUpdateDocument(inputs)

	/* Handle email change */
	var emailVerificationToken *string
	if email != nil {
		fmt.Println("SET EMAIL")
		temp := uuid.NewString()
		emailVerificationToken = &temp
		updateEntries = append(updateEntries, bson.E{Key: "pending_email", Value: email})
		updateEntries = append(updateEntries, bson.E{Key: "email_verification_token", Value: emailVerificationToken})
		updateEntries = append(updateEntries, bson.E{Key: "email_verification_token_expiry_date", Value: time.Now().Add(time.Minute * 60 * 24 * 1)})
	}

	/* Handle password change */
	if currentPassword != nil && newPassword != nil {
		fmt.Println("SET PASSWORD")

		if isCorrectPassword(user, *currentPassword) {
			fmt.Println("NEW_PASSWORD")
			updateEntries = append(updateEntries, bson.E{Key: "password_hash", Value: newPassword})
			response.ChangePassword = &UpdateState{Success: true}
		} else {
			fmt.Println("INVALID_CREDENTIALS")
			errorText := ErrorInvalidCredentials.Error()
			response.ChangePassword = &UpdateState{Success: false, Error: &errorText}
		}

	}

	fmt.Println("UPDATE_ENTRIES", updateEntries, emailVerificationToken)

	if len(updateEntries) > 0 {
		update := bson.D{{Key: "$set", Value: updateEntries}}
		fmt.Println("UPDATE", update)

		var document bson.D
		unmarshalError := userCollection.FindOneAndUpdate(dbContext, filter, update, options).Decode(&document)
		fmt.Println(unmarshalError, document)

		if unmarshalError != nil {
			return nil, unmarshalError
		}
	}

	if firstName != nil {
		response.FirstName = &UpdateState{Success: true}
	}

	if lastName != nil {
		response.LastName = &UpdateState{Success: true}
	}

	if city != nil {
		response.City = &UpdateState{Success: true}
	}

	if email != nil {
		if emailVerificationToken == nil {
			errorText := ErrorEmailVerificationToken.Error()
			response.EmailVerification = &UpdateState{Success: false, Error: &errorText}
		} else {
			emailError := SendEmailVerification(*email, *emailVerificationToken)
			if emailError == nil {
				response.EmailVerification = &UpdateState{Success: true}
			} else {
				errorText := emailError.Error()
				response.EmailVerification = &UpdateState{Success: false, Error: &errorText}
			}
		}
	}

	fmt.Println("RESPONSE", response)

	return &response, nil

}

func isCorrectPassword(user *User, passwordHash string) bool {
	return hash(user.PasswordSalt+user.PasswordHash) == passwordHash
}

func SendEmailVerification(email string, emailVerificationToken string) error {
	return sendMail(email, "E-Mail bestätigen", "<b>Klicke auf diesen Link, um deine E-Mail zu bestätigen:<b><br/><a href=\"http://localhost:3000/verify_email/"+emailVerificationToken+"\">Email bestätigen</a>")
}

func sendMail(email string, subject string, htmlBody string) error {

	mail := gomail.NewMessage()
	mail.SetHeader("From", "help@yacoid.de")
	mail.SetHeader("To", email)
	mail.SetHeader("Subject", subject)
	mail.SetBody("text/html", htmlBody)

	dialer := gomail.NewDialer("localhost", 2525, "email", "password")

	if err := dialer.DialAndSend(mail); err != nil {
		return err
	}

	return nil

}

func DeleteUser(authToken string, passwordHash string, reason string) error {

	fmt.Println(authToken, passwordHash, reason)
	user, findError := GetUserByAuthToken(authToken)

	if findError != nil {
		return findError
	}

	if isCorrectPassword(user, passwordHash) {

		filter := bson.M{"_id": user.ID}

		var result bson.D
		err := userCollection.FindOneAndDelete(dbContext, filter).Decode(&result)

		return err

	} else {
		return ErrorInvalidCredentials
	}

}
