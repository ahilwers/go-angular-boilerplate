package mongodb_test

import (
	"context"
	"testing"
	"time"

	"boilerplate/internal/entities"
	"boilerplate/internal/storage/mongodb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	testDBName = "testdb"
)

// setupMongoContainer sets up a MongoDB container for testing
func setupMongoContainer(t *testing.T) (string, func()) {
	ctx := context.Background()

	// Define the container request
	req := tc.ContainerRequest{
		Image:        "mongo:latest",
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor:   wait.ForLog("Waiting for connections").WithStartupTimeout(30 * time.Second),
	}

	// Start the container
	container, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err, "failed to start container")

	// Get the container host and port
	port, err := container.MappedPort(ctx, "27017")
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	// Create a connection string
	uri := "mongodb://" + host + ":" + port.Port()

	// Return the connection string and a cleanup function
	return uri, func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}
}

// createTestClient creates a MongoDB client for testing
func createTestClient(t *testing.T, uri string) *mongo.Client {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	require.NoError(t, err, "failed to connect to MongoDB")

	t.Cleanup(func() {
		if err := client.Disconnect(context.Background()); err != nil {
			t.Fatalf("failed to disconnect from MongoDB: %v", err)
		}
	})

	return client
}

func TestMongoDbProjectRepository_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Set up MongoDB container
	mongoURI, cleanup := setupMongoContainer(t)
	defer cleanup()

	// Create test client
	client := createTestClient(t, mongoURI)

	// Create repository
	repo := mongodb.NewProjectRepository(client, testDBName)

	// Test data - don't set ID, let Insert generate it
	testProject := &entities.Project{
		Name:        "Test Project",
		Description: "Test Description",
	}

	t.Run("Insert and FindByID", func(t *testing.T) {
		// Insert project
		err := repo.Insert(testProject)
		require.NoError(t, err, "failed to insert project")
		require.NotEmpty(t, testProject.ID, "expected ID to be set after insert")

		// Find project by ID
		found, err := repo.FindByID(testProject.ID)
		require.NoError(t, err, "failed to find project by ID")
		assert.Equal(t, testProject.ID, found.ID)
		assert.Equal(t, testProject.Name, found.Name)
		assert.Equal(t, testProject.Description, found.Description)
		assert.False(t, found.CreatedAt.IsZero(), "expected CreatedAt to be set")
		assert.False(t, found.UpdatedAt.IsZero(), "expected UpdatedAt to be set")
	})

	t.Run("Update", func(t *testing.T) {
		// Update project
		updatedName := "Updated Test Project"
		updatedDesc := "Updated Description"
		testProject.Name = updatedName
		testProject.Description = updatedDesc

		// Save the old UpdatedAt to verify it changes
		oldUpdatedAt := testProject.UpdatedAt

		// Small delay to ensure UpdatedAt changes
		time.Sleep(10 * time.Millisecond)

		err := repo.Update(testProject)
		require.NoError(t, err, "failed to update project")

		// Verify update
		found, err := repo.FindByID(testProject.ID)
		require.NoError(t, err)
		assert.Equal(t, updatedName, found.Name)
		assert.Equal(t, updatedDesc, found.Description)
		assert.True(t, found.UpdatedAt.After(oldUpdatedAt), "expected UpdatedAt to be updated")
	})

	t.Run("FindAll", func(t *testing.T) {
		// Find all projects
		projects, err := repo.FindAll()
		require.NoError(t, err, "failed to find all projects")
		assert.GreaterOrEqual(t, len(projects), 1, "expected at least one project")
	})

	t.Run("Delete", func(t *testing.T) {
		// Delete project
		err := repo.Delete(testProject.ID)
		require.NoError(t, err, "failed to delete project")

		// Verify deletion
		_, err = repo.FindByID(testProject.ID)
		require.Error(t, err, "expected error when finding deleted project")
	})

	t.Run("FindByID_NonExistent", func(t *testing.T) {
		// Try to find non-existent project
		nonExistentID := primitive.NewObjectID().Hex()
		_, err := repo.FindByID(nonExistentID)
		require.Error(t, err, "expected error for non-existent project")
	})

	t.Run("Update_NonExistent", func(t *testing.T) {
		// Try to update non-existent project
		nonExistentProject := &entities.Project{
			ID:          primitive.NewObjectID().Hex(),
			Name:        "Non-existent Project",
			Description: "This project does not exist",
		}
		err := repo.Update(nonExistentProject)
		require.Error(t, err, "expected error when updating non-existent project")
	})

	t.Run("Delete_NonExistent", func(t *testing.T) {
		// Try to delete non-existent project
		nonExistentID := primitive.NewObjectID().Hex()
		err := repo.Delete(nonExistentID)
		require.Error(t, err, "expected error when updating non-existent project")
	})
}
