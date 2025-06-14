package mongodb

import (
	"boilerplate/internal/entities"
	"boilerplate/internal/storage"
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoDbTask struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Name      string             `bson:"name"`
	ProjectID primitive.ObjectID `bson:"project_id,omitempty"`
}

func ToMongo(task entities.Task) (*MongoDbTask, error) {
	var oid primitive.ObjectID
	var err error

	if task.ID != "" {
		oid, err = primitive.ObjectIDFromHex(task.ID)
		if err != nil {
			return nil, err
		}
	} else {
		oid = primitive.NewObjectID()
	}

	var projectOid primitive.ObjectID
	if task.ProjectID != "" {
		projectOid, err = primitive.ObjectIDFromHex(task.ProjectID)
		if err != nil {
			return nil, errors.New("invalid project ID format")
		}
	}

	return &MongoDbTask{
		ID:        oid,
		Name:      task.Name,
		ProjectID: projectOid,
	}, nil
}

func FromMongo(task MongoDbTask) entities.Task {
	var projectID string
	if !task.ProjectID.IsZero() {
		projectID = task.ProjectID.Hex()
	}

	return entities.Task{
		ID:        task.ID.Hex(),
		Name:      task.Name,
		ProjectID: projectID,
	}
}

type TaskRepository struct {
	collection *mongo.Collection
	ctx        context.Context
}

func NewTaskRepository(client *mongo.Client, database string) storage.TaskRepository {
	collection := client.Database(database).Collection("tasks")
	return &TaskRepository{
		collection: collection,
		ctx:        context.Background(),
	}
}

func (r *TaskRepository) Insert(task *entities.Task) error {
	if task == nil {
		return errors.New("task cannot be nil")
	}

	if task.ID != "" {
		return errors.New("task already has an ID, use Update instead")
	}

	mongoTask := &MongoDbTask{
		ID:   primitive.NewObjectID(),
		Name: task.Name,
	}

	if task.ProjectID != "" {
		projectOid, err := primitive.ObjectIDFromHex(task.ProjectID)
		if err != nil {
			return errors.New("invalid project ID format")
		}
		mongoTask.ProjectID = projectOid
	}

	_, err := r.collection.InsertOne(r.ctx, mongoTask)
	if err != nil {
		return err
	}

	task.ID = mongoTask.ID.Hex()
	return nil
}

func (r *TaskRepository) Update(task *entities.Task) error {
	if task == nil {
		return errors.New("task cannot be nil")
	}

	if task.ID == "" {
		return errors.New("task has no ID, use Insert instead")
	}

	mongoTask, err := ToMongo(*task)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": mongoTask.ID}
	result, err := r.collection.ReplaceOne(r.ctx, filter, mongoTask)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("no task found with the given ID")
	}

	return nil
}

func (r *TaskRepository) Delete(id string) error {
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
		return errors.New("no task found with the given ID")
	}

	return nil
}

func (r *TaskRepository) FindByID(id string) (entities.Task, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return entities.Task{}, err
	}

	filter := bson.M{"_id": oid}
	var mongoTask MongoDbTask
	err = r.collection.FindOne(r.ctx, filter).Decode(&mongoTask)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return entities.Task{}, errors.New("task not found")
		}
		return entities.Task{}, err
	}

	return FromMongo(mongoTask), nil
}

func (r *TaskRepository) FindAll() ([]entities.Task, error) {
	cursor, err := r.collection.Find(r.ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(r.ctx)

	var mongoTasks []MongoDbTask
	if err := cursor.All(r.ctx, &mongoTasks); err != nil {
		return nil, err
	}

	tasks := make([]entities.Task, len(mongoTasks))
	for i, mongoTask := range mongoTasks {
		tasks[i] = FromMongo(mongoTask)
	}

	return tasks, nil
}

func (r *TaskRepository) FindByProjectID(projectID string) ([]entities.Task, error) {
	projectOid, err := primitive.ObjectIDFromHex(projectID)
	if err != nil {
		return nil, errors.New("invalid project ID format")
	}

	filter := bson.M{"project_id": projectOid}
	cursor, err := r.collection.Find(r.ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(r.ctx)

	var mongoTasks []MongoDbTask
	if err := cursor.All(r.ctx, &mongoTasks); err != nil {
		return nil, err
	}

	tasks := make([]entities.Task, len(mongoTasks))
	for i, mongoTask := range mongoTasks {
		tasks[i] = FromMongo(mongoTask)
	}

	return tasks, nil
}
