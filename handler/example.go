// internal/handler/example.go
package handler

import (
	"log"
	"fmt" // Added for potential parameter logging/debugging
	"github.com/gin-gonic/gin"
//	"dmmserver/game_error" // Assuming this is the correct path based on 30008.go
)

// init function to register the example handler.
// Replace "example_msg_id" with the actual message ID string.
func init() {
	Register("example_msg_id", handleExample)
}

// handleExample is a template for handling specific message IDs.
// It demonstrates parameter parsing, business logic placeholders,
// success response construction, and error handling.
//
// Parameters:
//   c *gin.Context: The Gin context for the request.
//   msgData map[string]interface{}: The data payload from the client message.
//
// Returns:
//   map[string]interface{}: A map containing the response data to be sent to the client.
//   error: An error object if any issue occurs, otherwise nil.
//          Use game_error.New(errorCode, errorMessage) for custom game errors.
func handleExample(c *gin.Context, msgData map[string]interface{}) (map[string]interface{}, error) {
	log.Printf("Executing handler for msg_id=example_msg_id. Ban checks (if applicable) should have passed in middleware.")
	log.Printf("Received msgData: %+v", msgData) // Log received data for debugging

	// 1. Parameter Parsing and Validation (Example)
	// ---------------------------------------------
	// Example: Retrieve a string parameter "playerName"
	// playerName, ok := msgData["playerName"].(string)
	// if !ok || playerName == "" {
	// 	log.Printf("Error: Missing or invalid 'playerName' parameter for msg_id=example_msg_id")
	// 	return nil, game_error.New(game_error.ErrorCodeBadRequest, "Missing or invalid 'playerName' parameter.")
	// }
	// log.Printf("Parsed playerName: %s", playerName)

	// Example: Retrieve an integer parameter "score"
	// scoreFloat, ok := msgData["score"].(float64) // JSON numbers are often float64
	// if !ok {
	// 	log.Printf("Error: Missing or invalid 'score' parameter for msg_id=example_msg_id")
	// 	return nil, game_error.New(game_error.ErrorCodeBadRequest, "Missing or invalid 'score' parameter.")
	// }
	// score := int(scoreFloat)
	// log.Printf("Parsed score: %d", score)


	// 2. Business Logic
	// -----------------
	// Implement your specific business logic here based on the parsed parameters.
	// This might involve:
	// - Interacting with other services (e.g., database, user service)
	// - Performing calculations
	// - Updating game state
	//
	// Example:
	// if playerName == "ForbiddenUser" {
	//   return nil, game_error.New(game_error.ErrorCodePermissionDenied, "User is not allowed to perform this action.")
	// }
	//
	// resultData, err := someBusinessOperation(playerName, score)
	// if err != nil {
	//   log.Printf("Error during business operation for msg_id=example_msg_id: %v", err)
	//   // You might want to return a more generic error to the client
	//   return nil, game_error.New(game_error.ErrorCodeInternalError, "An internal server error occurred.")
	// }


	// 3. Success Response Construction
	// --------------------------------
	// If the business logic is successful, construct the response data.
	// The structure of this map will be serialized to JSON and sent to the client.
	responseData := map[string]interface{}{
		"status":  "success",
		"message": fmt.Sprintf("Successfully processed example_msg_id with data: %+v", msgData),
		// "data": resultData, // Include actual result data from your business logic
		"example_key": "example_value",
	}

	log.Printf("Successfully processed msg_id=example_msg_id. Response: %+v", responseData)
	return responseData, nil

	// 4. Error Handling (Alternative to returning directly)
	// -----------------------------------------------------
	// If an error occurs at any point, you can return it.
	// The dispatcher or a higher-level handler will typically format this error
	// into a standardized error response for the client.
	//
	// Example of returning a custom game error:
	// return nil, game_error.New(game_error.ErrorCodeResourceNotFound, "The requested resource was not found.")
	//
	// Example of returning a generic error (less ideal for client-facing errors):
	// return nil, fmt.Errorf("an unexpected error occurred: %s", "details")
}

// Note: Ensure the Register function (likely in all.go or dispatcher.go)
// is correctly implemented to store and retrieve handlers by msg_id.
// The `init()` function above relies on `Register` being available in the `handler` package.