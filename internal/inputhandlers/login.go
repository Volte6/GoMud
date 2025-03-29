package inputhandlers

import (
	// ... other imports

	"github.com/volte6/gomud/internal/configs"
	"github.com/volte6/gomud/internal/connections"
	"github.com/volte6/gomud/internal/language"
	"github.com/volte6/gomud/internal/mudlog"
	"github.com/volte6/gomud/internal/term"
	"github.com/volte6/gomud/internal/users"
)

// Condition Helpers

func ConditionIsNewSignup(results map[string]string) bool {
	return results["new-signup"] == `new`
}

// ConditionUserDoesNotExist checks if the username entered does *not* exist.
func ConditionUserDoesNotExist(results map[string]string) bool {
	username := results["username"] // Assumes previous step had ID "username"
	return !users.Exists(username)
}

// ConditionUserExists checks if the username entered *does* exist.
func ConditionUserExists(results map[string]string) bool {
	username := results["username"] // Assumes previous step had ID "username"
	return users.Exists(username)
}

// FinalizeLoginOrCreate is called after all prompts are successfully answered.
func FinalizeLoginOrCreate(results map[string]string, sharedState map[string]any, clientInput *connections.ClientInput) bool {

	username := results["username"]
	password := results["password"]

	if username != `new` {
		userExists := users.Exists(username)

		if userExists {
			// Existing User Login Logic (No changes needed)
			tmpUser, err := users.LoadUser(username)
			if err != nil {
				mudlog.Error("Failed to load existing user during login", "username", username, "error", err)
				connections.SendTo([]byte(language.T("Error.LoginFailedGeneric")), clientInput.ConnectionId)
				connections.SendTo(term.CRLF, clientInput.ConnectionId)
				connections.Remove(clientInput.ConnectionId)
				return false // Indicate failure, connection removed
			}

			if !tmpUser.PasswordMatches(password) {
				connections.SendTo([]byte(`Nope. Bye!`), clientInput.ConnectionId)
				connections.SendTo(term.CRLF, clientInput.ConnectionId)
				connections.Remove(clientInput.ConnectionId)
				return false // Indicate failure, connection removed
			}

			loggedInUser, msg, err := users.LoginUser(tmpUser, clientInput.ConnectionId)
			if err != nil {
				connections.SendTo([]byte(msg), clientInput.ConnectionId)
				connections.SendTo(term.CRLF, clientInput.ConnectionId)
				connections.Remove(clientInput.ConnectionId)
				return false // Indicate failure, connection removed
			}

			sharedState["UserObject"] = loggedInUser // For main loop

			if len(msg) > 0 {
				connections.SendTo([]byte(msg), clientInput.ConnectionId)
				connections.SendTo(term.CRLF, clientInput.ConnectionId)
			}
			mudlog.Info("User logged in", "username", username, "connectionId", clientInput.ConnectionId)
			return true // Indicate success, handler can be removed

		} else {
			connections.SendTo([]byte(`Invalid login.`), clientInput.ConnectionId)
			connections.SendTo(term.CRLF, clientInput.ConnectionId)
			connections.Remove(clientInput.ConnectionId)
			return false // Indicate failure, connection removed
		}
	} else {
		/*
			username-new
			password-new
			password-new-verify
			email-new
			screen-reader-new y/n
			confirm_create y/n
		*/

		confirmCreate, exists := results["confirm_create"] // Assumes step ID "confirm_create"
		if !exists || confirmCreate != "y" {
			connections.SendTo([]byte(`Okay, bye!`), clientInput.ConnectionId) // Use language key
			connections.SendTo(term.CRLF, clientInput.ConnectionId)
			connections.Remove(clientInput.ConnectionId)
			return false // Indicate failure, connection removed
		}

		username := results["username-new"]
		password := results["password-new"]

		if users.Exists(results["username-new"]) {
			connections.SendTo([]byte(`I'm sorry, that user already exists!`), clientInput.ConnectionId) // Use language key
			connections.SendTo(term.CRLF, clientInput.ConnectionId)
			connections.Remove(clientInput.ConnectionId)
			return false
		}

		newUser := users.NewUserRecord(0, clientInput.ConnectionId)
		newUser.EmailAddress = results["email-new"]
		newUser.ScreenReader = results["screen-reader-new"] == `y`

		// Error handling for SetUsername/SetPassword might be redundant if validation passed, but good practice
		if err := newUser.SetUsername(username); err != nil {
			mudlog.Error("Internal error setting username post-validation", "username", username, "error", err)
			connections.SendTo([]byte(language.T("Error.UserCreationFailed")), clientInput.ConnectionId) // Generic creation error
			connections.SendTo(term.CRLF, clientInput.ConnectionId)
			connections.Remove(clientInput.ConnectionId)
			return false
		}
		if err := newUser.SetPassword(password); err != nil {
			mudlog.Error("Internal error setting password post-validation", "username", username, "error", err)
			connections.SendTo([]byte(language.T("Error.UserCreationFailed")), clientInput.ConnectionId) // Generic creation error
			connections.SendTo(term.CRLF, clientInput.ConnectionId)
			connections.Remove(clientInput.ConnectionId)
			return false
		}

		if err := users.CreateUser(newUser); err != nil {
			mudlog.Error("Could not create user", "username", username, "error", err)
			// Try to give specific feedback if possible, otherwise generic
			connections.SendTo([]byte(err.Error()), clientInput.ConnectionId)
			connections.SendTo(term.CRLF, clientInput.ConnectionId)
			connections.Remove(clientInput.ConnectionId)
			return false // Indicate failure, connection removed
		}

		sharedState["UserObject"] = newUser // For main loop

		mudlog.Info("New user created", "username", username, "connectionId", clientInput.ConnectionId)

		return true // Indicate success, handler can be removed
	}
}

func GetLoginPromptHandler() connections.InputHandler {

	// Define the steps for the login process
	loginSteps := []*PromptStep{
		{
			ID:             "username",
			PromptTemplate: "login/username.prompt",
			MaskInput:      false,
			Validator:      ValidateNewEntry,
		},
		//////////////////////////////////////////////////
		// If NOT a new user signup (Just a login)
		//////////////////////////////////////////////////
		{
			ID:             "password",
			PromptTemplate: "login/password.prompt",
			MaskInput:      true,
			MaskTemplate:   "login/password.mask", // Optional: specify if different from "*"
			Validator:      ValidatePassword,
			Condition:      func(results map[string]string) bool { return results["username"] != `new` }, // Only run if username was not "new"
		},
		//////////////////////////////////////////////////
		// End If NOT a new user signup (Just a login)
		//////////////////////////////////////////////////
		//////////////////////////////////////////////////
		// If a new user signup
		//////////////////////////////////////////////////
		{
			ID:             "username-new",
			PromptTemplate: "login/username-new.prompt",
			MaskInput:      false,
			Validator:      ValidateUsername,
			Condition:      func(results map[string]string) bool { return results["username"] == `new` }, // Only run if username was "new"
		},
		{
			ID:             "password-new",
			PromptTemplate: "login/password-new.prompt",
			MaskInput:      true,
			MaskTemplate:   "login/password.mask", // Optional: specify if different from "*"
			Validator:      ValidatePassword,
			Condition:      func(results map[string]string) bool { return results["username"] == `new` }, // Only run if username was "new"
		},
		{
			ID:             "password-new-verify",
			PromptTemplate: "login/password-new-verify.prompt",
			MaskInput:      true,
			MaskTemplate:   "login/password.mask", // Optional: specify if different from "*"
			Validator:      ValidatePassword2,
			Condition:      func(results map[string]string) bool { return results["username"] == `new` }, // Only run if username was "new"
		},
		{
			ID:             "email-new",
			PromptTemplate: "login/email-new.prompt",
			GetDataFunc: func(results map[string]string) map[string]any {
				// Dynamically generate the data for the generic y/n prompt
				return map[string]any{
					"emailIsOptional": configs.GetValidationConfig().EmailOnJoin != `required`,
				}
			},
			MaskInput: false,
			Validator: ValidateEmail,
			Condition: func(results map[string]string) bool {
				return results["username"] == `new` && configs.GetValidationConfig().EmailOnJoin != `none` // Only run if username was "new" and email is enabled
			},
		},
		{
			ID:             "screen-reader-new",
			PromptTemplate: "generic/prompt.yn",
			GetDataFunc: func(results map[string]string) map[string]any {
				// Dynamically generate the data for the generic y/n prompt
				return map[string]any{
					"prompt":  "Are you using a screen reader?",
					"options": []string{"y", "n"},
					"default": "n", // Default shown in the prompt, actual default on empty input handled by validator
				}
			},
			MaskInput: false,
			Validator: ValidateYesNo,
			Condition: func(results map[string]string) bool { return results["username"] == `new` }, // Only run if username was "new"
		},
		{
			ID:             "confirm_create",
			PromptTemplate: "generic/prompt.yn", // Use the generic yes/no template
			GetDataFunc: func(results map[string]string) map[string]any {
				// Dynamically generate the data for the generic y/n prompt
				return map[string]any{
					"prompt": language.T("Login.CreateUser", map[any]any{ // Use language.T for the prompt text
						"Username": results["username-new"], // Inject username
					}),
					"options": []string{"y", "n"},
					"default": "n", // Default shown in the prompt, actual default on empty input handled by validator
				}
			},
			MaskInput: false,
			Validator: ValidateYesNo,
			Condition: func(results map[string]string) bool { return results["username"] == `new` }, // Only run if username was "new"
		},
		//////////////////////////////////////////////////
		// End If a new user signup
		//////////////////////////////////////////////////
	}

	// Create and return the handler using the generic factory function
	return CreatePromptHandler(loginSteps, FinalizeLoginOrCreate)
}
