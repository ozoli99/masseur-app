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
    http.Handle("/", http.FileServer(http.Dir("./static")))
    
    http.HandleFunc("/appointments", appointmentsHandler)
    http.HandleFunc("/appointments/", appointmentHandler)

    fmt.Println("Server running on http://localhost:8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func appointmentsHandler(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
        case http.MethodGet:
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(appointments)
        case http.MethodPost:
            var a Appointment
            if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
                http.Error(w, err.Error(), http.StatusBadRequest)
                return
            }
            a.ID = nextID
            nextID++
            if a.Time.IsZero() {
                a.Time = time.Now()
            }
            appointments = append(appointments, a)
            w.WriteHeader(http.StatusCreated)
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

    idx := -1
    for i, ap := range appointments {
        if ap.ID == id {
            idx = i
            break
        }
    }

    if idx == -1 {
        http.Error(w, "Not Found", http.StatusNotFound)
        return
    }

    switch r.Method {
        case http.MethodGet:
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(appointments[idx])
        case http.MethodPut:
            var updated Appointment
            if err := json.NewDecoder(r.Body).Decode(&updated); err != nil {
                http.Error(w, err.Error(), http.StatusBadRequest)
                return
            }
            updated.ID = appointments[idx].ID
            appointments[idx] = updated
            w.Header().Set("Content-Type", "application/json")
            json.NewEncoder(w).Encode(updated)
        case http.MethodDelete:
            appointments = append(appointments[:idx], appointments[idx+1:]...)
            w.WriteHeader(http.StatusNoContent)
        default:
            http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
    }
}