package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func main() {
    if err := initDatabase(); err != nil {
        log.Fatalf("Error initializing database: %v", err)
    }

    http.Handle("/", http.FileServer(http.Dir("./static")))
    
    http.HandleFunc("/appointments", appointmentsHandler)
    http.HandleFunc("/appointments/", appointmentHandler)

    fmt.Println("Server running on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func appointmentsHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
        case http.MethodGet:
            queryValues := r.URL.Query()
            startStr := queryValues.Get("start")
            endStr := queryValues.Get("end")
            customer := queryValues.Get("customer")
            sortField := queryValues.Get("sort")

            baseQuery := "SELECT id, customer_name, time, duration, notes FROM appointments"
            var conditions []string
            var args []interface{}

            if startStr != "" && endStr != "" {
                conditions = append(conditions, "time BETWEEN ? AND ?")
                args = append(args, startStr, endStr)
            }

            if customer != "" {
                conditions = append(conditions, "customer_name LIKE ?")
                args = append(args, "%"+customer+"%")
            }

            if len(conditions) > 0 {
                baseQuery += " WHERE " + strings.Join(conditions, " AND ")
            }

            allowedSorts := map[string]string {
                "time": "time",
                "customer_name": "customer_name",
                "duration": "duration",
            }
            if sortField != "" {
                if col, ok := allowedSorts[sortField]; ok {
                    baseQuery += " ORDER BY " + col
                }
            }

            rows, err := database.Query(baseQuery, args...)
            if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
            defer rows.Close()

            var result []Appointment
            for rows.Next() {
                var a Appointment
                var t string
                if err := rows.Scan(&a.ID, &a.CustomerName, &t, &a.Duration, &a.Notes); err != nil {
                    http.Error(w, err.Error(), http.StatusInternalServerError)
                    return
                }

                parsedTime, err := time.Parse(time.RFC3339, t)
                if err == nil {
                    a.Time = parsedTime
                } else {
                    a.Time = time.Now()
                }
                
                result = append(result, a)
            }
        
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(result)
        case http.MethodPost:
            var a Appointment
            if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
                http.Error(w, err.Error(), http.StatusBadRequest)
                return
            }
            if a.Time.IsZero() {
                a.Time = time.Now()
            }

            res, err := database.Exec("INSERT INTO appointments (customer_name, time, duration, notes) VALUES (?, ?, ?, ?)",a.CustomerName, a.Time.Format(time.RFC3339), a.Duration, a.Notes)
            if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }

            id, err := res.LastInsertId()
            if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
            a.ID = int(id)

            w.WriteHeader(http.StatusCreated)
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(a)
        default:
            http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
    }
}

func appointmentHandler(w http.ResponseWriter, r *http.Request) {
    parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
    if len(parts) < 2 {
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    idStr := parts[1]
    id, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, "Invalid appointment ID", http.StatusBadRequest)
        return
    }

    switch r.Method {
        case http.MethodGet:
            var a Appointment
            var t string
            err := database.QueryRow("SELECT id, customer_name, time, duration, notes FROM appointments WHERE id = ?", id).Scan(&a.ID, &a.CustomerName, &t, &a.Duration, &a.Notes)
            if err != nil {
                http.Error(w, "Not Found", http.StatusNotFound)
            }
            aTime, err := time.Parse(time.RFC3339, t)
            if err == nil {
                a.Time = aTime
            }

            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(a)
        case http.MethodPut:
            var updated Appointment
            if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
                http.Error(w, err.Error(), http.StatusBadRequest)
                return
            }
            if updated.Time.IsZero() {
                updated.Time = time.Now()
            }

            _, err := database.Exec("UPDATE appointments SET customer_name = ?, time = ?, duration = ?, notes = ? WHERE id = ?", updated.CustomerName, updated.Time.Format(time.RFC3339), updated.Duration, updated.Notes, id)
            if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }

            updated.ID = id
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(updated)
        case http.MethodDelete:
            _, err := database.Exec("DELETE FROM appointments WHERE id = ?", id)
            if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
            }
            w.WriteHeader(http.StatusNoContent)
        default:
            http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
    }
}