package data

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/mongocollectibles/rental-system/models"
)

// DynamoDBRepository implements Repository interface using DynamoDB
type DynamoDBRepository struct {
	client            *dynamodb.Client
	collectiblesTable string
	rentalsTable      string
	warehousesTable   string
}

// NewDynamoDBRepository creates a new DynamoDB repository
func NewDynamoDBRepository(cfg aws.Config) *DynamoDBRepository {
	return &DynamoDBRepository{
		client:            dynamodb.NewFromConfig(cfg),
		collectiblesTable: "MongoCollectibles-Collectibles",
		rentalsTable:      "MongoCollectibles-Rentals",
		warehousesTable:   "MongoCollectibles-Warehouses",
	}
}

// GetAllCollectibles returns all collectibles
func (r *DynamoDBRepository) GetAllCollectibles() ([]*models.Collectible, error) {
	out, err := r.client.Scan(context.TODO(), &dynamodb.ScanInput{
		TableName: aws.String(r.collectiblesTable),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scan collectibles: %w", err)
	}

	var collectibles []*models.Collectible
	if err := attributevalue.UnmarshalListOfMaps(out.Items, &collectibles); err != nil {
		return nil, fmt.Errorf("failed to unmarshal collectibles: %w", err)
	}

	return collectibles, nil
}

// GetCollectibleByID returns a collectible by ID
func (r *DynamoDBRepository) GetCollectibleByID(id string) (*models.Collectible, error) {
	out, err := r.client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(r.collectiblesTable),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get collectible: %w", err)
	}
	if out.Item == nil {
		return nil, fmt.Errorf("collectible not found")
	}

	var collectible models.Collectible
	if err := attributevalue.UnmarshalMap(out.Item, &collectible); err != nil {
		return nil, fmt.Errorf("failed to unmarshal collectible: %w", err)
	}
	return &collectible, nil
}

// AddCollectible adds a new collectible
func (r *DynamoDBRepository) AddCollectible(collectible *models.Collectible) error {
	item, err := attributevalue.MarshalMap(collectible)
	if err != nil {
		return fmt.Errorf("failed to marshal collectible: %w", err)
	}

	_, err = r.client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(r.collectiblesTable),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to put collectible: %w", err)
	}
	return nil
}

// GetWarehouses returns warehouses for a collectible
func (r *DynamoDBRepository) GetWarehouses(collectibleID string) ([]models.Warehouse, error) {
	out, err := r.client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String(r.warehousesTable),
		KeyConditionExpression: aws.String("collectible_id = :cid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":cid": &types.AttributeValueMemberS{Value: collectibleID},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query warehouses: %w", err)
	}

	var warehouses []models.Warehouse
	if err := attributevalue.UnmarshalListOfMaps(out.Items, &warehouses); err != nil {
		return nil, fmt.Errorf("failed to unmarshal warehouses: %w", err)
	}
	return warehouses, nil
}

// AddWarehouse adds a warehouse for a collectible
func (r *DynamoDBRepository) AddWarehouse(collectibleID string, warehouse models.Warehouse) error {
	// Ensure warehouse struct has collectible_id set, might be redundant but safe
	warehouse.CollectibleID = collectibleID

	item, err := attributevalue.MarshalMap(warehouse)
	if err != nil {
		return err
	}
	// Set warehouse_id = id
	item["warehouse_id"] = item["id"]

	_, err = r.client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(r.warehousesTable),
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("failed to put warehouse: %w", err)
	}
	return nil
}

func (r *DynamoDBRepository) GetAllWarehouses() (map[string][]models.Warehouse, error) {
	out, err := r.client.Scan(context.TODO(), &dynamodb.ScanInput{
		TableName: aws.String(r.warehousesTable),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scan warehouses: %w", err)
	}

	var allWarehouses []models.Warehouse
	if err := attributevalue.UnmarshalListOfMaps(out.Items, &allWarehouses); err != nil {
		return nil, fmt.Errorf("failed to unmarshal warehouses: %w", err)
	}

	result := make(map[string][]models.Warehouse)
	for _, w := range allWarehouses {
		result[w.CollectibleID] = append(result[w.CollectibleID], w)
	}
	return result, nil
}

// CreateRental creates a new rental record
func (r *DynamoDBRepository) CreateRental(rental *models.Rental) error {
	// Ensure GSI key is populated
	if rental.CustomerEmail == "" {
		rental.CustomerEmail = rental.Customer.Email
	}

	item, err := attributevalue.MarshalMap(rental)
	if err != nil {
		return fmt.Errorf("failed to marshal rental: %w", err)
	}

	_, err = r.client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(r.rentalsTable),
		Item:      item,
		// ConditionExpression: aws.String("attribute_not_exists(id)"), // Ensure uniqueness
	})
	if err != nil {
		return fmt.Errorf("failed to create rental: %w", err)
	}
	return nil
}

// GetRentalByID returns a rental by ID
func (r *DynamoDBRepository) GetRentalByID(id string) (*models.Rental, error) {
	out, err := r.client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(r.rentalsTable),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get rental: %w", err)
	}
	if out.Item == nil {
		return nil, fmt.Errorf("rental not found")
	}

	var rental models.Rental
	if err := attributevalue.UnmarshalMap(out.Item, &rental); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rental: %w", err)
	}
	return &rental, nil
}

// UpdateRental updates an existing rental
func (r *DynamoDBRepository) UpdateRental(rental *models.Rental) error {
	// For simplicity, just PutItem (overwrite)
	return r.CreateRental(rental)
}

// GetAllRentals scans all rentals
func (r *DynamoDBRepository) GetAllRentals() ([]*models.Rental, error) {
	out, err := r.client.Scan(context.TODO(), &dynamodb.ScanInput{
		TableName: aws.String(r.rentalsTable),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scan rentals: %w", err)
	}

	var rentals []*models.Rental
	if err := attributevalue.UnmarshalListOfMaps(out.Items, &rentals); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rentals: %w", err)
	}
	return rentals, nil
}

// GetRentalsByCustomerAndCollectible queries using GSI or Scan
func (r *DynamoDBRepository) GetRentalsByCustomerAndCollectible(email string, collectibleID string) ([]*models.Rental, error) {
	// If we have a GSI on email, we can Query that, then filter by collectibleID
	// Template says: GSI CustomerEmailIndex

	out, err := r.client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              aws.String(r.rentalsTable),
		IndexName:              aws.String("CustomerEmailIndex"),
		KeyConditionExpression: aws.String("customer_email = :email"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":email": &types.AttributeValueMemberS{Value: email},
			":cid":   &types.AttributeValueMemberS{Value: collectibleID},
		},
		FilterExpression: aws.String("collectible_id = :cid"),
	})
	if err != nil {
		log.Printf("DynamoDB Query failed: %v", err)
		return nil, err
	}

	var rentals []*models.Rental
	if err := attributevalue.UnmarshalListOfMaps(out.Items, &rentals); err != nil {
		return nil, err
	}
	return rentals, nil
}

// DeleteAllRentals clears all rental records in DynamoDB
func (r *DynamoDBRepository) DeleteAllRentals() error {
	// 1. Scan all rentals to get keys
	out, err := r.client.Scan(context.TODO(), &dynamodb.ScanInput{
		TableName:            aws.String(r.rentalsTable),
		ProjectionExpression: aws.String("id"), // Only fetch keys
	})
	if err != nil {
		return fmt.Errorf("failed to scan for deletion: %w", err)
	}

	if len(out.Items) == 0 {
		return nil
	}

	log.Printf("Deleting %d rentals...", len(out.Items))

	// 2. Delete item by item (BatchWriteItem is more efficient but limit 25 items, loop needed. PutItem loop is simpler for now)
	for _, item := range out.Items {
		var key struct {
			ID string `dynamodbav:"id"`
		}
		if err := attributevalue.UnmarshalMap(item, &key); err != nil {
			log.Printf("Failed to unmarshal key: %v", err)
			continue
		}

		_, err := r.client.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
			TableName: aws.String(r.rentalsTable),
			Key: map[string]types.AttributeValue{
				"id": &types.AttributeValueMemberS{Value: key.ID},
			},
		})
		if err != nil {
			log.Printf("Failed to delete rental %s: %v", key.ID, err)
			// Continue deleting others even if one fails
		}
	}

	return nil
}
