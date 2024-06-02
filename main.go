package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"


	// "os"
	"path/filepath"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/google/uuid"
	"github.com/jcelliott/lumber"
)

const(
	Version       = "1.0.1"
	AWS_S3_REGION = ""
	AWS_S3_BUCKET = ""
	AWS_S3_PREFIX = ""

)


var sess = connectAWS()

func connectAWS() *session.Session {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(AWS_S3_REGION),
	})
	if err != nil {
		fmt.Println(err)
	}
	return sess
}

func uploadStuffs(sess *session.Session, bucket, key string, content []byte) error {

	r := bytes.NewReader(content)

	uploader := s3manager.NewUploader(sess)

	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   r,
	})
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func downloadFromS3(sess *session.Session, bucket, key string) ([]byte, error) {
	downloader := s3manager.NewDownloader(sess)

	buf := aws.NewWriteAtBuffer([]byte{})
	_, err := downloader.Download(buf, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	return buf.Bytes(), nil
}

func deleteFromS3(sess *session.Session, bucket, key string) error {
	svc := s3.New(sess)
	_, err := svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func listFromS3(sess *session.Session, bucket, prefix string) ([]string, error) {
	svc := s3.New(sess)
	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	var keys []string
	for _, item := range resp.Contents {
		keys = append(keys, *item.Key)
	}

	return keys, nil
}

func upload(w http.ResponseWriter, r *http.Request) {
	uploadStuffs(sess, AWS_S3_BUCKET, AWS_S3_PREFIX+"/Jyoti.json", []byte(`{"Name": "Jyoti", "Age": 22, "Contact": "1234567890", "Company": "ABC", "Address": {"City": "Bhubaneshwar", "State": "Odisha", "Country": "India", "Pincode": 751024}}`))
	w.Write([]byte("Uploaded"))
}

type (
	Logger interface {
		Fatal(string, ...interface{})
		Error(string, ...interface{})
		Warn(string, ...interface{})
		Info(string, ...interface{})
		Debug(string, ...interface{})
		Trace(string, ...interface{})
	}
	Driver struct {
		mutex   sync.Mutex
		mutexes map[string]*sync.Mutex
		// dir     string
		log Logger
	}
)

type Options struct {
	Logger
}

func New(options *Options) (*Driver, error) {
	// dir = filepath.Clean(dir)

	opts := Options{}

	if options != nil {
		opts = *options
	}
	if opts.Logger == nil {
		opts.Logger = lumber.NewConsoleLogger(lumber.INFO)
	}

	driver := Driver{
		mutexes: make(map[string]*sync.Mutex),
		log:     opts.Logger,
	}

	return &driver, nil
}

func (d *Driver) Write(collection, resource string, v interface{}) error {
	if collection == "" {
		return fmt.Errorf("collection is required")
	}
	if resource == "" {
		return fmt.Errorf("resource is required")
	}
	mutex := d.getOrCreateMutex(collection)
	mutex.Lock()
	defer mutex.Unlock()

	// dir := filepath.Join(d.dir, collection)
	// key := filepath.Join(AWS_S3_PREFIX, collection, resource+".json")
	key := fmt.Sprintf("%s/%s/%s.json", AWS_S3_PREFIX, collection, resource)
	// fnlPath := filepath.Join(dir, resource+".json")
	// tempPath := fnlPath + ".tmp"
	// if err := os.MkdirAll(dir, 0755); err != nil {
	// 	return err
	// }

	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return err
	}

	b = append(b, byte('\n'))

	// if err := os.WriteFile(tempPath, b, 0644); err != nil {
	// 	return err
	// }

	// return os.Rename(tempPath, fnlPath)

	return uploadStuffs(sess, AWS_S3_BUCKET, key, b)

}

func (d *Driver) Read(collection, resource string, v interface{}) error {
	if collection == "" {
		return fmt.Errorf("missing collection")
	}
	if resource == "" {
		return fmt.Errorf("missing Resource")
	}
	// record := filepath.Join(d.dir, collection, resource+".json")
	// key := filepath.Join(AWS_S3_PREFIX, collection, resource+".json")
	key := fmt.Sprintf("%s/%s/%s.json", AWS_S3_PREFIX, collection, resource)

	// if _, err := stat(record); err != nil {
	// 	return err
	// }
	b, err := downloadFromS3(sess, AWS_S3_BUCKET, key)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, &v)
}

func (d *Driver) ReadAll(collection string) ([]string, error) {
	if collection == "" {
		return nil, fmt.Errorf("missing collection")
	}
	// dir := filepath.Join(d.dir, collection)
	// prefix := filepath.Join(AWS_S3_PREFIX, collection)
	prefix := fmt.Sprintf("%s/%s", AWS_S3_PREFIX, collection)

	keys, err := listFromS3(sess, AWS_S3_BUCKET, prefix)
	// if err != nil {

	if err != nil {
		return nil, err
	}

	// files, _ := os.ReadDir(dir)

	var records []string

	for _, key := range keys {
		// b, err := os.ReadFile(filepath.Join(dir, files.Name()))
		b, err := downloadFromS3(sess, AWS_S3_BUCKET, key)
		if err != nil {
			return nil, err
		}
		records = append(records, string(b))
	}
	return records, nil
}

func (d *Driver) Delete(collection, resource string) error {
	// path := filepath.Join(d.dir, collection, resource)
	// key := filepath.Join(AWS_S3_PREFIX,  collection, resource+".json")
	key := fmt.Sprintf("%s/%s/%s.json", AWS_S3_PREFIX, collection, resource)
	mutex := d.getOrCreateMutex(collection)
	mutex.Lock()
	defer mutex.Unlock()

	// dir := filepath.Join(d.dir, path)

	// switch fi, err := stat(dir); {
	// case fi == nil, err != nil:
	// 	return fmt.Errorf("unable to find file or dir name %v", path)

	// case fi.Mode().IsDir():
	// 	return os.RemoveAll(dir)

	// case fi.Mode().IsRegular():
	// 	return os.RemoveAll(dir + ".json")
	// }

	return deleteFromS3(sess, AWS_S3_BUCKET, key)
}

func (d *Driver) DeleteAll(collection string) error {
	// prefix := filepath.Join(AWS_S3_PREFIX, collection)
	prefix := filepath.Join(AWS_S3_PREFIX, collection)
	keys, err := listFromS3(sess, AWS_S3_BUCKET, prefix)
	if err != nil {
		return err
	}

	for _, key := range keys {
		if err := deleteFromS3(sess, AWS_S3_BUCKET, key); err != nil {
			return err
		}
	}

	return nil
}

func (d *Driver) getOrCreateMutex(collection string) *sync.Mutex {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	m, ok := d.mutexes[collection]

	if !ok {
		m = &sync.Mutex{}
		d.mutexes[collection] = m
	}
	return m
}

// func stat(path string) (fi os.FileInfo, err error) {
// 	if fi, err = os.Stat(path); os.IsNotExist(err) {
// 		fi, err = os.Stat(path + ".json")
// 	}
// 	return

// }

type Address struct {
	City    string
	State   string
	Country string
	Pincode json.Number
}

type User struct {
	ID      string
	Name    string
	Age     json.Number
	Contact string
	Company string
	Address Address
	Number  string
}




func main() {

	
	db, err := New(nil)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	employees := []User{
		{ID: uuid.New().String(), Name: "Jyoti", Age: "22", Contact: "1234567890", Company: "ABC", Address: Address{City: "Bhubaneshwar", State: "Odisha", Country: "India", Pincode: "751024"}},
		{ID: uuid.New().String(), Name: "Lol", Age: "22", Contact: "1234567890", Company: "ABC", Address: Address{City: "Mumbai", State: "Odisha", Country: "India", Pincode: "751024"}},
		{ID: uuid.New().String(), Name: "Tar", Age: "22", Contact: "1234567890", Company: "ABC", Address: Address{City: "USA", State: "Odisha", Country: "India", Pincode: "751024"}},
		{ID: uuid.New().String(), Name: "Sam", Age: "22", Contact: "1234567890", Company: "ABC", Address: Address{City: "Kerela", State: "Odisha", Country: "India", Pincode: "751024"}},
		{ID: uuid.New().String(), Name: "Ram", Age: "22", Contact: "1234567890", Company: "ABC", Address: Address{City: "UP", State: "Odisha", Country: "India", Pincode: "751024"}},
		{ID: uuid.New().String(), Name: "Sol", Age: "22", Contact: "1234567890", Company: "ABC", Address: Address{City: "Kolkata", State: "Odisha", Country: "India", Pincode: "751024"}},
	}

	for _, value := range employees {
		db.Write("users", value.Name, User{
			ID:      value.ID,
			Name:    value.Name,
			Age:     value.Age,
			Contact: value.Contact,
			Company: value.Company,
			Address: value.Address,
		})
	}

	records, err := db.ReadAll("users")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}
	fmt.Println(records)

	allusers := []User{}
	for _, f := range records {
		employeeFound := User{}
		if err := json.Unmarshal([]byte(f), &employeeFound); err != nil {
			fmt.Printf("Error: %s\n", err)
		}
		allusers = append(allusers, employeeFound)
	}
	fmt.Println((allusers))

	// connectAWS()

	// r := mux.NewRouter()

	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		// This is a demo upload endpoint. You can modify as needed.
		demoUser := User{
			ID:      uuid.New().String(),
			Name:    "Demo",
			Age:     "30",
			Contact: "0987654321",
			Company: "XYZ",
			Address: Address{City: "DemoCity", State: "DemoState", Country: "DemoCountry", Pincode: "000000"},
		}
		err := db.Write("users", demoUser.ID, demoUser)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Error uploading user: %s", err)))
			return
		}
		w.Write([]byte("Uploaded"))
	})

	// uploadFile(sess, AWS_S3_BUCKET, AWS_S3_PREFIX, "./users/Jyoti.json")

	// r.HandleFunc("/upload", upload).Methods("POST")

	http.ListenAndServe(":3000", nil)
	// if err:= db.Delete("users", "Jyoti"); err!=nil{
	// 	fmt.Printf("Error: %s\n", err)
	// }

	// if err:= db.Delete("users", ""); err!=nil{
	// 	fmt.Printf("Error: %s\n", err)
	// }
}
