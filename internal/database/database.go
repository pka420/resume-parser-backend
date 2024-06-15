package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
	"resume-backend-parser/internal/models"
    "errors"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
    pq "github.com/lib/pq" // Import PostgreSQL driver

)

// Service represents a service that interacts with a database.
type Service interface {
	Health() map[string]string
    CreateUser(name string, email string, password_hash string, address string, profile_headline string) error
    UserExists(email string) (bool, error)
    CheckUserLogin(email string, password string) (bool, error)
    IsUserAdmin(email string) (bool, error)
    GetUserId(email string) (int, error)
    UpdateProfile(userId int, resumeFileAddress string) error
    UpdateProfileWithFields(userId int, profile models.ProfileThirdParty) error

    CreateJob(title string, description string, companyName string, TotalApplications int, userId int) error
    GetJob(id int) (models.Job, error)
    GetJobs() ([]models.Job, error)
    GetApplicants(jobId int) ([]models.Profile, error)

    ApplyJob(jobId int, userId int) error

    GetApplicantProfile(userId int) (models.Profile, error)
    GetAllApplicants() ([]models.Profile, error)


	Close() error
}

type service struct {
	db *sql.DB
}

var (
	database   = os.Getenv("DB_DATABASE")
	password   = os.Getenv("DB_PASSWORD")
	username   = os.Getenv("DB_USERNAME")
	port       = os.Getenv("DB_PORT")
	host       = os.Getenv("DB_HOST")
	schema     = os.Getenv("DB_SCHEMA")
	dbInstance *service
)

func New() Service {
    connStr := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
        host, port, username, password, database, "",
    )

    db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	dbInstance = &service{
		db: db,
	}
	return dbInstance
}

func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)
	err := s.db.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf(fmt.Sprintf("db down: %v", err)) // Log the error and terminate the program
		return stats
	}

	stats["status"] = "up"
	stats["message"] = "It's healthy"

	dbStats := s.db.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	if dbStats.OpenConnections > 40 { // Assuming 50 is the max for this example
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

func (s *service) Close() error {
	log.Printf("Disconnected from database: %s", database)
	return s.db.Close()
}

// func (s *service) GetUser(id int) string {
//     query := "SELECT * FROM users where id = 1"
// }


func (s *service) CreateUser(name string, email string, password_hash string, address string, profile_headline string) error {
    query := "INSERT INTO users (name, email, password_hash, address, profile_headline, type) VALUES ($1, $2, $3, $4, $5, $6)"
    _, err := s.db.Exec(query, name, email, password_hash, address, profile_headline, "user")
    return err
}

func (s *service) UserExists(email string) (bool, error) {
    query := "SELECT id FROM users WHERE email = $1"
    row := s.db.QueryRow(query, email)
    var emailExists string
    err := row.Scan(&emailExists)
    if errors.Is(err, sql.ErrNoRows) {
        return false, nil
    } else if err != nil {
        return false, err
    }
    return true, nil
}

func (s *service) CheckUserLogin(email string, password string) (bool, error) {
    query := "SELECT id FROM users WHERE email = $1 AND password_hash = $2"
    row := s.db.QueryRow(query, email, password)
    var emailExists string
    err := row.Scan(&emailExists)
    if err != nil {
        return false, err
    }
    return true, nil
}

func (s *service) CreateJob(title string, description string, companyName string, TotalApplications int, userId int) error {
    now := time.Now()
    emtpyArray := sql.NullInt64{}
    query := "INSERT INTO jobs (title, description, company_name, total_applications, applicants, posted_by, posted_on) VALUES ($1, $2, $3, $4, $5, $6, $7)"
    _, err := s.db.Exec(query, title, description, companyName, TotalApplications, emtpyArray, userId, now)
    return err
}

func (s *service) GetJobs() ([]models.Job, error) {
    query := "SELECT id FROM jobs"
    rows, err := s.db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var jobs []models.Job
    var jobId int
    for rows.Next() {
        err = rows.Scan(&jobId)
        if err != nil {
            return jobs, err
        }
        fmt.Println(jobId)
        var job models.Job
        var applicants []sql.NullInt64
        query := "SELECT id, title, description, posted_on, total_applications, posted_by, company_name, applicants FROM jobs WHERE id = $1"
        row := s.db.QueryRow(query, jobId)
        err := row.Scan(&job.Id, &job.Title, &job.Description,
            &job.PostedOn, &job.TotalApplications, &job.PostedBy,
            &job.CompanyName, pq.Array(&applicants))
        if err != nil {
            return nil, err
        }
        intApplicants := make([]int, len(applicants))
        for i, a := range applicants {
            if a.Valid {
                intApplicants[i] = int(a.Int64) // Ensure it's safe to convert
            }
        }
        job.Applicants = intApplicants
        jobs = append(jobs, job)
    }
    return jobs, nil
}

func (s *service) IsUserAdmin(email string) (bool, error) {
    query := "SELECT id FROM users WHERE email = $1 AND type = 'admin'"
    row := s.db.QueryRow(query, email)
    var emailExists string
    err := row.Scan(&emailExists)
    if errors.Is(err, sql.ErrNoRows) {
        return false, nil
    } else if err != nil {
        return false, err
    }
    return true, nil
}

func (s *service) GetUserId(email string) (int, error) {
    query := "SELECT id FROM users WHERE email = $1"
    row := s.db.QueryRow(query, email)
    var id int
    err := row.Scan(&id)
    if err != nil {
        return 0, err
    }
    return id, nil
}

func (s *service) GetJob(id int) (models.Job, error) {
    query := "SELECT id, title, description, posted_on, total_applications, posted_by, company_name, applicants FROM jobs WHERE id = $1"
    row := s.db.QueryRow(query, id)
    var job models.Job
    var applicants []sql.NullInt64
    err := row.Scan(&job.Id, &job.Title, &job.Description,
        &job.PostedOn, &job.TotalApplications, &job.PostedBy,
        &job.CompanyName, pq.Array(&applicants))
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return models.Job{}, errors.New("Job not found")
        }
        return models.Job{}, err
    }
    intApplicants := make([]int, len(applicants))
    for i, a := range applicants {
        if a.Valid {
            intApplicants[i] = int(a.Int64) // Ensure it's safe to convert
        }
    }
    job.Applicants = intApplicants
    return job, nil
}

func (s *service) GetApplicantProfile(userId int) (models.Profile, error) {
    query := "SELECT applicant, skills, education, experience, name, email, phone FROM profile WHERE applicant = $1"
    row := s.db.QueryRow(query, userId)
    var profile models.Profile
    err := row.Scan(&profile.Applicant, &profile.ResumeFileAddress, &profile.Skills, &profile.Education, &profile.Experience, &profile.Name, &profile.Email, &profile.Phone)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return models.Profile{}, errors.New("Profile not found")
        }
        return models.Profile{}, err
    }
    return profile, nil
}

func (s *service) GetApplicants(jobId int) ([]models.Profile, error) {
    var applicants []sql.NullInt64
    query := "SELECT applicants FROM jobs WHERE id = $1"
    row := s.db.QueryRow(query, jobId)
    err := row.Scan(pq.Array(&applicants))
    if err != nil {
        return nil, err
    }
    var profiles []models.Profile
    for _, a := range applicants {
        if a.Valid {
            query := "SELECT applicant, skills, education, experience, name, email, phone  FROM profile WHERE applicant = $1"
            row := s.db.QueryRow(query, a.Int64)
            var profile models.Profile
            err := row.Scan(&profile.Applicant, &profile.ResumeFileAddress, &profile.Skills, &profile.Education, &profile.Experience, &profile.Name, &profile.Email, &profile.Phone)
            if err != nil {
                return nil, err
            }
            profiles = append(profiles, profile)
        }
    }
    return profiles, nil
}

func (s *service) GetAllApplicants() ([]models.Profile, error) {
    query := "SELECT applicant, skills, education, experience, name, email, phone FROM profile"
    rows, err := s.db.Query(query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var profiles []models.Profile
    for rows.Next() {
        var profile models.Profile
        err = rows.Scan(&profile.Applicant, &profile.ResumeFileAddress, &profile.Skills, &profile.Education, &profile.Experience, &profile.Name, &profile.Email, &profile.Phone)
        if err != nil {
            return profiles, err
        }
        profiles = append(profiles, profile)
    }
    return profiles, nil
}

func (s *service) UpdateProfile(userId int, resumeFileAddress string) error {
    query := "SELECT applicant FROM profile WHERE applicant = $1"
    row := s.db.QueryRow(query, userId)
    var id int
    err := row.Scan(&id)
    if errors.Is(err, sql.ErrNoRows) {
        query = "INSERT INTO profile (applicant, resume_file_address) VALUES ($1, $2)"
        _, err = s.db.Exec(query, userId, resumeFileAddress)
        return err
    } else if err != nil {
        return err
    }
    query = "UPDATE profile set resume_file_address = $1 WHERE applicant = $2"
    _, err = s.db.Exec(query, resumeFileAddress, userId)
    return err

}

func (s *service) ApplyJob(jobId int, userId int) error {
    //check job exists
    query := "SELECT id FROM jobs WHERE id = $1"
    row := s.db.QueryRow(query, jobId)
    var id int
    err := row.Scan(&id)
    if err != nil {
        return err
    }

    query = "UPDATE jobs SET applicants = array_append(applicants, $1) WHERE id = $2"
    _, err = s.db.Exec(query, userId, jobId)
    if err != nil {
        return err
    }
    return err
}
func (s *service) UpdateProfileWithFields(userId int, profile models.ProfileThirdParty) error {
    fmt.Println("update profile with fields", userId)
    query := "SELECT applicant FROM profile WHERE applicant = $1"
    row := s.db.QueryRow(query, userId)
    var id int
    err := row.Scan(&id)
    if err != nil {
        return err
    }
    query = "UPDATE profile set name, email, phone, education, experience, skills = $1, $2, $3, $4, $5, $6 WHERE applicant = $7"
    _, err = s.db.Exec(query, profile.Name, profile.Email, profile.Phone, profile.Education, profile.Experience, pq.Array(profile.Skills), userId)
    return err
}
