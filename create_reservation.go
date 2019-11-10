package garderobel

import (
	"cloud.google.com/go/firestore"
	"context"
	"encoding/json"
	firebase "firebase.google.com/go"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// FirestoreEvent is the payload of a Firestore event.
type FirestoreEvent struct {
	OldValue   FirestoreValue `json:"oldValue"`
	Value      FirestoreValue `json:"value"`
	UpdateMask struct {
		FieldPaths []string `json:"fieldPaths"`
	} `json:"updateMask"`
}

// FirestoreValue holds Firestore fields.
type FirestoreValue struct {
	CreateTime time.Time `json:"createTime"`
	// Fields is the data for this value. The type depends on the format of your
	// database. Log an interface{} value and inspect the result to see a JSON
	// representation of your database fields.
	Fields     MyData    `json:"fields"`
	Name       string    `json:"name"`
	UpdateTime time.Time `json:"updateTime"`
}

type Request struct {
	VenueId    string `json:"venueId"`
	WardrobeId string `json:"wardrobeId"`
	SectionId  string `json:"sectionId"`
	UserId     string `json:"userId"`
}

// MyData represents a value from Firestore. The type definition depends on the
// format of your database.
type MyData struct {
	Original struct {
		StringValue string `json:"stringValue"`
	} `json:"original"`
}

func findAvailableHanger(ref *firestore.DocumentRef) *firestore.DocumentSnapshot {
	iter := ref.Collection(`hangers`).Where(`state`, `==`, 0).Limit(1).Documents(context.Background())
	list, err := iter.GetAll()
	if err != nil || len(list) == 0 {
		return nil
	}
	return list[0]

}

func respond(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	fmt.Fprint(w, msg)
}

//noinspection GoUnusedExportedFunction
func CreateReservation(w http.ResponseWriter, r *http.Request) {
	log.Print("test")
	fmt.Print("test")
	var params Request
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		respond(w, 401, err.Error())
		return
	}

	venueRef := client.Collection(`venues`).Doc(params.VenueId)
	log.Print(params)

	wardrobeRef := venueRef.Collection(`wardrobes`).Doc(params.WardrobeId)

	sectionRef := wardrobeRef.Collection(`sections`).Doc(params.SectionId)

	hanger := findAvailableHanger(sectionRef)
	if hanger == nil {
		respond(w, 404, "No available hangers")
		return
	}

	hangerData := map[string]interface{}{"state": 1, "stateUpdated": firestore.ServerTimestamp}

	_, err := hanger.Ref.Set(context.Background(), hangerData, firestore.MergeAll)

	userRef := client.Collection(`users`).Doc(params.UserId)

	hangerName, err := hanger.DataAt(`id`)
	if err != nil {
		return
	}
	user, err := userRef.Get(context.Background())

	userName, err := user.DataAt(`name`)

	venue, err := venueRef.Get(context.Background())
	if err != nil {
		respond(w, 401, err.Error())
	}
	venueName, err := venue.DataAt(`name`)

	//wardrobe, err := wardrobeRef.Get(context.Background())
	//if err != nil {
	//	w.WriteHeader(401)
	//	fmt.Fprint(w, "Invalid wardrobeRef ID")
	//	return
	//}
	//
	//section, err := sectionRef.Get(context.Background())
	//if err != nil {
	//	w.WriteHeader(401)
	//	fmt.Fprint(w, "Invalid sectionRef ID")
	//	return
	//}
	reservationData := map[string]interface{}{
		`section`:         sectionRef,
		`hanger`:          hanger.Ref,
		`hangerName`:      hangerName,
		`user`:            userRef,
		`userName`:        userName,
		`venue`:           venueRef,
		`venueName`:       venueName,
		`wardrobe`:        wardrobeRef,
		`state`:           4,
		`reservationTime`: firestore.ServerTimestamp,
	}

	ref, _, err := client.Collection(`reservations`).Add(context.Background(), reservationData)
	if ref == nil {
		w.WriteHeader(401)
		fmt.Fprint(w, err.Error())
		return
	}
	fmt.Fprint(w, ref.Path)

}

// GCLOUD_PROJECT is automatically set by the Cloud Functions runtime.
var projectID = os.Getenv("GCLOUD_PROJECT")

// client is a Firestore client, reused between function invocations.
var client *firestore.Client

func init() {
	// Use the application default credentials.
	conf := &firebase.Config{ProjectID: projectID}

	// Use context.Background() because the app/client should persist across
	// invocations.
	ctx := context.Background()

	app, err := firebase.NewApp(ctx, conf)
	if err != nil {
		log.Fatalf("firebase.NewApp: %v", err)
	}

	client, err = app.Firestore(ctx)
	if err != nil {
		log.Fatalf("app.Firestore: %v", err)
	}
}
