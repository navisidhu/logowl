package services

import (
	"errors"
	"regexp"
	"time"

	"github.com/navisidhu/logowl/internal/keys"
	"github.com/navisidhu/logowl/internal/models"
	"github.com/navisidhu/logowl/internal/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type InterfaceOrganization interface {
	CheckPresence(bson.M) (bool, error)
	Create(models.Organization) (primitive.ObjectID, error)
	Delete(primitive.ObjectID) error
	FindOne(bson.M) (models.Organization, error)
	FindOneAndUpdate(bson.M, bson.M) (models.Organization, error)
}

type Organization struct {
	Store store.InterfaceStore
}

func (o *Organization) CheckPresence(filter bson.M) (bool, error) {
	return o.Store.Organization().CheckPresence(filter)
}

func (o *Organization) Create(organization models.Organization) (primitive.ObjectID, error) {
	timestamp := time.Now()

	organization.MonthlyRequestLimit = keys.GetKeys().MONTHLY_REQUEST_LIMIT
	organization.Plan = "free"
	organization.SubscriptionID = ""
	organization.IsSetUp = false
	organization.CreatedAt = timestamp
	organization.UpdatedAt = timestamp

	if keys.GetKeys().IS_SELFHOSTED {
		organization.IsSetUp = true
	}

	if !organization.Validate() {
		return primitive.NilObjectID, errors.New("the provided organization data is invalid")
	}

	regex := regexp.MustCompile(`\s+`)
	organization.Identifier = regex.ReplaceAllString(organization.Name, "")

	return o.Store.Organization().InsertOne(organization)
}

func (o *Organization) Delete(organizationID primitive.ObjectID) error {
	allServices, err := o.Store.Service().Find(bson.M{"organizationId": organizationID})
	if err != nil {
		return err
	}

	var allServiceIDs []primitive.ObjectID
	var allTickets []string

	for _, service := range allServices {
		allServiceIDs = append(allServiceIDs, service.ID)
		allTickets = append(allTickets, service.Ticket)
	}

	c := make(chan error, 5)

	go func() {
		if len(allServiceIDs) == 0 {
			c <- nil
			return
		}

		_, err := o.Store.Service().DeleteMany(bson.M{"_id": bson.M{"$in": allServiceIDs}})
		c <- err
	}()

	go func() {
		if len(allTickets) == 0 {
			c <- nil
			return
		}

		_, err := o.Store.Error().DeleteMany(bson.M{"ticket": bson.M{"$in": allTickets}})
		c <- err
	}()

	go func() {
		if len(allTickets) == 0 {
			c <- nil
			return
		}

		_, err := o.Store.Analytics().DeleteMany(bson.M{"ticket": bson.M{"$in": allTickets}})
		c <- err
	}()

	go func() {
		_, err := o.Store.Organization().DeleteOne(bson.M{"_id": organizationID})
		c <- err
	}()

	go func() {
		_, err := o.Store.User().DeleteMany(bson.M{"organizationId": organizationID})
		c <- err
	}()

	var failed error

	for i := 0; i < 4; i++ {
		err := <-c

		if err != nil {
			failed = err
		}
	}

	return failed
}

func (o *Organization) FindOne(filter bson.M) (models.Organization, error) {
	return o.Store.Organization().FindOne(filter)
}

func (o *Organization) FindOneAndUpdate(filter, update bson.M) (models.Organization, error) {
	update["updatedAt"] = time.Now()

	return o.Store.Organization().FindOneAndUpdate(filter, bson.M{"$set": update})
}

func GetOrganizationService(store store.InterfaceStore) Organization {
	return Organization{store}
}
