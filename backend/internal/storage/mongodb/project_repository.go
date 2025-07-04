package mongodb

import (
	"boilerplate/internal/entities"
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongoDbProject struct {
	ID   primitive.ObjectID `bson:"_id,omitempty"`
	Name string             `bson:"name"`
}

type mongoDbProjectRepository struct {
	collection *mongo.Collection
	ctx        context.Context
}

func NewProjectRepository(client *mongo.Client, database string) *mongoDbProjectRepository {
	collection := client.Database(database).Collection("projects")
	return &mongoDbProjectRepository{
		collection: collection,
		ctx:        context.Background(),
	}
}

func (r *mongoDbProjectRepository) Insert(project *entities.Project) error {
	if project == nil {
		return errors.New("project cannot be nil")
	}

	if project.ID != "" {
		return errors.New("project already has an ID, use Update instead")
	}

	mongoProject := &mongoDbProject{
		ID:   primitive.NewObjectID(),
		Name: project.Name,
	}

	_, err := r.collection.InsertOne(r.ctx, mongoProject)
	if err != nil {
		return err
	}

	project.ID = mongoProject.ID.Hex()
	return nil
}

func (r *mongoDbProjectRepository) Update(project *entities.Project) error {
	if project == nil {
		return errors.New("project cannot be nil")
	}

	if project.ID == "" {
		return errors.New("project has no ID, use Insert instead")
	}

	mongoProject, err := toMongoProject(*project)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": mongoProject.ID}
	result, err := r.collection.ReplaceOne(r.ctx, filter, mongoProject)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("no project found with the given ID")
	}

	return nil
}

func (r *mongoDbProjectRepository) Delete(id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": oid}
	result, err := r.collection.DeleteOne(r.ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("no project found with the given ID")
	}

	return nil
}

func (r *mongoDbProjectRepository) FindByID(id string) (entities.Project, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return entities.Project{}, err
	}

	filter := bson.M{"_id": oid}
	var mongoProject mongoDbProject
	err = r.collection.FindOne(r.ctx, filter).Decode(&mongoProject)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return entities.Project{}, errors.New("project not found")
		}
		return entities.Project{}, err
	}

	return fromMongoProject(mongoProject), nil
}

func (r *mongoDbProjectRepository) FindAll() ([]entities.Project, error) {
	cursor, err := r.collection.Find(r.ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(r.ctx)

	var mongoProjects []mongoDbProject
	if err := cursor.All(r.ctx, &mongoProjects); err != nil {
		return nil, err
	}

	projects := make([]entities.Project, len(mongoProjects))
	for i, mongoProject := range mongoProjects {
		projects[i] = fromMongoProject(mongoProject)
	}

	return projects, nil
}

func (r *mongoDbProjectRepository) FindByProjectID(projectID string) (entities.Project, error) {
	return r.FindByID(projectID)
}

func toMongoProject(project entities.Project) (*mongoDbProject, error) {
	var oid primitive.ObjectID
	var err error

	if project.ID != "" {
		oid, err = primitive.ObjectIDFromHex(project.ID)
		if err != nil {
			return nil, err
		}
	} else {
		oid = primitive.NewObjectID()
	}

	return &mongoDbProject{
		ID:   oid,
		Name: project.Name,
	}, nil
}

func fromMongoProject(project mongoDbProject) entities.Project {
	return entities.Project{
		ID:   project.ID.Hex(),
		Name: project.Name,
	}
}
