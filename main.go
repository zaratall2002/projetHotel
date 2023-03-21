package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

//Structure du client

type Client struct {
	ID            int    `json:"id"`
	Nom           string `json:"nom"`
	Prenom        string `json:"prenom"`
	Telephone     string `json:"telephone"`
	Email         string `json:"email"`
	ClasseChambre string `json:"classeChambre"`
}

// Connexion à la base de données MySQL

func dbConnect() (*sql.DB, error) {

	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/hotel")
	if err != nil {
		return nil, err
	}
	return db, nil
}

func getClientsHandler(w http.ResponseWriter, r *http.Request) {
	// Connexion à la base de données
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/hotel")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Erreur lors de la connexion à la base de données")
		return
	}
	defer db.Close()

	// Récupération de tous les clients depuis la table "client"
	rows, err := db.Query("SELECT id, nom, prenom, telephone, email,classeChambre FROM client")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Erreur lors de la récupération des clients depuis la base de données")
		return
	}
	defer rows.Close()

	// Boucle sur les résultats et stockage dans une slice de clients
	clients := make([]Client, 0)
	for rows.Next() {
		var client Client
		err := rows.Scan(&client.ID, &client.Nom, &client.Prenom, &client.Telephone, &client.Email, &client.ClasseChambre)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "Erreur lors de la lecture des données du client depuis la base de données")
			return
		}
		clients = append(clients, client)
	}
	if err := rows.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Erreur lors de la récupération des clients depuis la base de données")
		return
	}

	// Envoi des clients en format JSON dans la réponse HTTP
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(clients)
}

// Création d'un nouveau client
func createClientHandler(w http.ResponseWriter, r *http.Request) {
	// Lecture des données du client à partir du corps de la requête
	var client Client
	err := json.NewDecoder(r.Body).Decode(&client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Connexion à la base de données
	db, err := dbConnect()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Exécution de la requête d'insertion
	result, err := db.Exec("INSERT INTO client(nom, prenom,telephone,email,classeChambre) VALUES (?, ?, ?, ?,?)", client.Nom, client.Prenom, client.Telephone, client.Email,
		client.ClasseChambre)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Récupération de l'ID généré automatiquement
	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Création de la réponse
	client.ID = int(id)
	response, err := json.Marshal(client)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Envoi de la réponse
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}
func updateClientHandler(w http.ResponseWriter, r *http.Request) {
	// Récupère l'ID du client à mettre à jour depuis les paramètres de la requête
	clientIDStr := r.URL.Query().Get("id")
	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	// Récupère les données du client à partir du corps de la requête
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var updatedClient Client
	err = json.Unmarshal(reqBody, &updatedClient)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	// Ouvre une connexion à la base de données MySQL
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/hotel")
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Prépare et exécute la requête de mise à jour du client dans la base de données
	query := "UPDATE client SET nom=?, prenom=?, telephone=?, email=?,classeChambre=? WHERE id=?"
	_, err = db.Exec(query, updatedClient.Nom, updatedClient.Prenom, updatedClient.Telephone, updatedClient.Email, updatedClient.ClasseChambre, clientID)
	if err != nil {
		http.Error(w, "Failed to update client in database", http.StatusInternalServerError)
		return
	}

	// Retourne une réponse HTTP 200 OK
	w.WriteHeader(http.StatusOK)
}
func deleteClientHandler(w http.ResponseWriter, r *http.Request) {
	// Récupérer l'id du client à supprimer à partir de la requête URL
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	// Ouvrir la connexion à la base de données
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/hotel")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Préparer la requête SQL pour supprimer le client
	stmt, err := db.Prepare("DELETE FROM client WHERE id = ?")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	// Exécuter la requête SQL pour supprimer le client
	res, err := stmt.Exec(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Vérifier si le client a été supprimé avec succès
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "Client not found", http.StatusNotFound)
		return
	}

	// Renvoyer une réponse JSON pour indiquer que le client a été supprimé avec succès
	response := map[string]interface{}{
		"status":  "success",
		"message": "Client deleted successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
func getClient(clientID int) (Client, error) {
	// Ouvre une connexion à la base de données MySQL
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/hotel")
	if err != nil {
		return Client{}, err
	}
	defer db.Close()

	// Prépare et exécute la requête de récupération du client depuis la base de données
	query := "SELECT nom, prenom, telephone, email,classeChambre FROM client WHERE id=?"
	row := db.QueryRow(query, clientID)

	// Récupère les données du client depuis la ligne de résultat
	var client Client
	err = row.Scan(&client.Nom, &client.Prenom, &client.Telephone, &client.Email, &client.ClasseChambre)
	if err != nil {
		if err == sql.ErrNoRows {
			return Client{}, fmt.Errorf("client with ID %d not found", clientID)
		}
		return Client{}, err
	}

	return client, nil
}

func getClientHandler(w http.ResponseWriter, r *http.Request) {
	// Récupère l'ID du client depuis l'URL
	clientID, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/clients/"))
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	// Récupère le client correspondant depuis la base de données
	client, err := getClient(clientID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Encode le client en JSON et renvoie la réponse HTTP
	json.NewEncoder(w).Encode(client)
}

// Structure reservation

type Reservation struct {
	ID          int    `json:"id"`
	Nom         string `json:"nom"`
	Prenom      string `json:"prenom"`
	Telephone   string `json:"telephone"`
	Classe      string `json:"classe"`
	Chambre     int    `json:"chambre"`
	DateChambre string `json:"DateChambre"`
	Nuite       int    `json:"Nuite"`
	DateSortie  string `json:"DateSortie"`
}

func dbConnection() (*sql.DB, error) {

	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/hotel")
	if err != nil {
		return nil, err
	}
	return db, nil
}

func getReservationsHandler(w http.ResponseWriter, r *http.Request) {
	// Connexion à la base de données
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/hotel")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Erreur lors de la connexion à la base de données")
		return
	}
	defer db.Close()

	// Récupération de tous les reservations depuis la table "reservation"
	rows, err := db.Query("SELECT id,nom,prenom,telephone,classe,chambre,DateChambre,Nuite,DateSortie FROM reservation")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Erreur lors de la récupération des reservations depuis la base de données")
		return
	}
	defer rows.Close()

	// Boucle sur les résultats et stockage dans une slice de reservation
	reservations := make([]Reservation, 0)
	for rows.Next() {
		var reservation Reservation
		err := rows.Scan(&reservation.ID, &reservation.Nom, &reservation.Prenom, &reservation.Telephone, &reservation.Classe, &reservation.Chambre, &reservation.DateChambre, &reservation.Nuite, &reservation.DateSortie)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "Erreur lors de la lecture des données de la reservation depuis la base de données")
			return
		}
		reservations = append(reservations, reservation)
	}
	if err := rows.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Erreur lors de la récupération des reservations depuis la base de données")
		return
	}

	// Envoi des clients en format JSON dans la réponse HTTP
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(reservations)
}
func createReservationHandler(w http.ResponseWriter, r *http.Request) {
	// Lecture des données du client à partir du corps de la requête
	var reservation Reservation
	err := json.NewDecoder(r.Body).Decode(&reservation)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Connexion à la base de données
	db, err := dbConnection()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Exécution de la requête d'insertion
	result, err := db.Exec("INSERT INTO reservation(nom,prenom,telephone,classe,chambre,DateChambre,Nuite,DateSortie) VALUES (?, ?, ?, ?,?,?,?,?)", reservation.Nom, reservation.Prenom, reservation.Telephone, reservation.Classe,
		reservation.Chambre, reservation.DateChambre, reservation.Nuite, reservation.DateSortie)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Récupération de l'ID généré automatiquement
	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Création de la réponse
	reservation.ID = int(id)
	response, err := json.Marshal(reservation)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Envoi de la réponse
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}

func updateReservationHandler(w http.ResponseWriter, r *http.Request) {
	// Récupère l'ID du client à mettre à jour depuis les paramètres de la requête
	reservationIDStr := r.URL.Query().Get("id")
	reservationID, err := strconv.Atoi(reservationIDStr)
	if err != nil {
		http.Error(w, "Invalid reservation ID", http.StatusBadRequest)
		return
	}

	// Récupère les données du client à partir du corps de la requête
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var updatedReservation Reservation
	err = json.Unmarshal(reqBody, &updatedReservation)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	// Ouvre une connexion à la base de données MySQL
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/hotel")
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Prépare et exécute la requête de mise à jour du client dans la base de données
	query := "UPDATE reservation SET nom=?, prenom=?,telephone=?,classe=?,chambre=?,DateChambre=?,Nuite=?,DateSortie=?  WHERE id=?"
	_, err = db.Exec(query, updatedReservation.Nom, updatedReservation.Prenom, updatedReservation.Telephone, updatedReservation.Classe, updatedReservation.Chambre, updatedReservation.DateChambre, updatedReservation.Nuite, updatedReservation.DateSortie, reservationID)
	if err != nil {
		http.Error(w, "Failed to update client in database", http.StatusInternalServerError)
		return
	}

	// Retourne une réponse HTTP 200 OK
	w.WriteHeader(http.StatusOK)
}
func deleteReservationHandler(w http.ResponseWriter, r *http.Request) {
	// Récupérer l'id du client à supprimer à partir de la requête URL
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	// Ouvrir la connexion à la base de données
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/hotel")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Préparer la requête SQL pour supprimer le client
	stmt, err := db.Prepare("DELETE FROM reservation WHERE id = ?")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	// Exécuter la requête SQL pour supprimer le client
	res, err := stmt.Exec(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Vérifier si le client a été supprimé avec succès
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "Reservation not found", http.StatusNotFound)
		return
	}

	// Renvoyer une réponse JSON pour indiquer que le client a été supprimé avec succès
	response := map[string]interface{}{
		"status":  "success",
		"message": "Reservation deleted successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
func getReservation(reservationID int) (Reservation, error) {
	// Ouvre une connexion à la base de données MySQL
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/hotel")
	if err != nil {
		return Reservation{}, err
	}
	defer db.Close()

	// Prépare et exécute la requête de récupération du client depuis la base de données
	query := "SELECT nom,prenom,telephone,classe,chambre,DateChambre,Nuite,DateSOrtie FROM reservation WHERE id=?"
	row := db.QueryRow(query, reservationID)

	// Récupère les données du client depuis la ligne de résultat
	var reservation Reservation
	err = row.Scan(&reservation.Nom, &reservation.Prenom, &reservation.Telephone, &reservation.Classe, &reservation.Chambre, &reservation.DateChambre, &reservation.Nuite, &reservation.DateSortie)
	if err != nil {
		if err == sql.ErrNoRows {
			return Reservation{}, fmt.Errorf("reservation with ID %d not found", reservationID)
		}
		return Reservation{}, err
	}

	return reservation, nil
}
func getReservationHandler(w http.ResponseWriter, r *http.Request) {
	// Récupère l'ID du client depuis l'URL
	reservationID, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/reservations/"))
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	// Récupère le client correspondant depuis la base de données
	reservation, err := getReservation(reservationID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Encode le client en JSON et renvoie la réponse HTTP
	json.NewEncoder(w).Encode(reservation)
}

type Chambre struct {
	ID            int    `json:"id"`
	Numero        int    `json:"numero"`
	Etage         int    `json:"etage"`
	Disponibilite string `json:"disponiblilite"`
	TypeChambre   string `json:"TypeChambre"`
	PrixParNuit   int    `json:"PrixParNuit"`
}

func dbConnexion() (*sql.DB, error) {

	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/hotel")
	if err != nil {
		return nil, err
	}
	return db, nil
}
func getChambresHandler(w http.ResponseWriter, r *http.Request) {
	// Connexion à la base de données
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/hotel")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Erreur lors de la connexion à la base de données")
		return
	}
	defer db.Close()

	// Récupération de toutes les chambres depuis la table "chambre"
	rows, err := db.Query("SELECT id,numero, etage,disponibilite,TypeChambre,PrixParNuit FROM chambre")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Erreur lors de la récupération des chambres depuis la base de données")
		return
	}
	defer rows.Close()

	// Boucle sur les résultats et stockage dans une slice de reservation
	chambres := make([]Chambre, 0)
	for rows.Next() {
		var chambre Chambre
		err := rows.Scan(&chambre.ID, &chambre.Numero, &chambre.Etage, &chambre.Disponibilite, &chambre.TypeChambre, &chambre.PrixParNuit)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "Erreur lors de la lecture des données  depuis la base de données")
			return
		}
		chambres = append(chambres, chambre)
	}
	if err := rows.Err(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Erreur lors de la récupération des données depuis la base de données")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(chambres)
}
func createChambreHandler(w http.ResponseWriter, r *http.Request) {
	// Lecture des données du client à partir du corps de la requête
	var chambre Chambre
	err := json.NewDecoder(r.Body).Decode(&chambre)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Connexion à la base de données
	db, err := dbConnexion()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Exécution de la requête d'insertion
	result, err := db.Exec("INSERT INTO chambre(numero, etage, disponibilite,TypeChambre,PrixParNuit) VALUES (?, ?, ?, ?,?)", chambre.Numero, chambre.Etage, chambre.Disponibilite,
		chambre.TypeChambre, chambre.PrixParNuit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Récupération de l'ID généré automatiquement
	id, err := result.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Création de la réponse
	chambre.ID = int(id)
	response, err := json.Marshal(chambre)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Envoi de la réponse
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(response)
}

func updateChambreHandler(w http.ResponseWriter, r *http.Request) {
	// Récupère l'ID du client à mettre à jour depuis les paramètres de la requête
	chambreIDStr := r.URL.Query().Get("id")
	chambreID, err := strconv.Atoi(chambreIDStr)
	if err != nil {
		http.Error(w, "Invalid Room ID", http.StatusBadRequest)
		return
	}

	// Récupère les données du client à partir du corps de la requête
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var updatedChambre Chambre
	err = json.Unmarshal(reqBody, &updatedChambre)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	// Ouvre une connexion à la base de données MySQL
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/hotel")
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Prépare et exécute la requête de mise à jour du client dans la base de données
	query := "UPDATE chambre SET numero=?,etage=?,disponibilite=?,TypeChambre=?,PrixParNuit=? WHERE id=?"
	_, err = db.Exec(query, updatedChambre.Numero, updatedChambre.Etage, updatedChambre.Disponibilite, updatedChambre.TypeChambre, updatedChambre.PrixParNuit, chambreID)
	if err != nil {
		http.Error(w, "Failed to update client in database", http.StatusInternalServerError)
		return
	}

	// Retourne une réponse HTTP 200 OK
	w.WriteHeader(http.StatusOK)
}
func deleteChambreHandler(w http.ResponseWriter, r *http.Request) {
	// Récupérer l'id du client à supprimer à partir de la requête URL
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	// Ouvrir la connexion à la base de données
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/hotel")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Préparer la requête SQL pour supprimer le client
	stmt, err := db.Prepare("DELETE FROM chambre WHERE id = ?")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	// Exécuter la requête SQL pour supprimer le client
	res, err := stmt.Exec(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Vérifier si le client a été supprimé avec succès
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rowsAffected == 0 {
		http.Error(w, "Reservation not found", http.StatusNotFound)
		return
	}

	// Renvoyer une réponse JSON pour indiquer que le client a été supprimé avec succès
	response := map[string]interface{}{
		"status":  "success",
		"message": "Room deleted successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
func getChambre(chambreID int) (Chambre, error) {
	// Ouvre une connexion à la base de données MySQL
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/hotel")
	if err != nil {
		return Chambre{}, err
	}
	defer db.Close()

	// Prépare et exécute la requête de récupération de la chambrz depuis la base de données
	query := "SELECT numero,etage,disponibilite,TypeChambre,PrixParNuit FROM chambre WHERE id=?"
	row := db.QueryRow(query, chambreID)

	// Récupère les données du client depuis la ligne de résultat
	var chambre Chambre
	err = row.Scan(&chambre.Numero, &chambre.Etage, &chambre.Disponibilite, &chambre.TypeChambre, &chambre.PrixParNuit)
	if err != nil {
		if err == sql.ErrNoRows {
			return Chambre{}, fmt.Errorf("room with ID %d not found", chambreID)
		}
		return Chambre{}, err
	}

	return chambre, nil
}
func getChambreHandler(w http.ResponseWriter, r *http.Request) {
	// Récupère l'ID du client depuis l'URL
	chambreID, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/chambres/"))
	if err != nil {
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	// Récupère le client correspondant depuis la base de données
	chambre, err := getChambre(chambreID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Encode le client en JSON et renvoie la réponse HTTP
	json.NewEncoder(w).Encode(chambre)
}

// Statistiques
type Statistique struct {
	ID                  int    `json:"id"`
	NbreChambreReservee int    `json:"NbreChambreReservee"`
	NbreChambreLibre    int    `json:"NbreChambreLibre"`
	TauxOccupation      string `json:"TauxOccupation"`
	NomHotel            string `json:"NomHotel"`
	AdresseHotel        string `json:"AdressHotel"`
}

func getStatistique(statistiqueID int) (Statistique, error) {
	// Ouvre une connexion à la base de données MySQL
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/hotel")
	if err != nil {
		return Statistique{}, err
	}
	defer db.Close()

	// Prépare et exécute la requête de récupération de la chambrz depuis la base de données
	query := "SELECT NbreChambreReservee,NbreChambreLibre,TauxOccupation,NomHotel,AdressHotel FROM statistique WHERE id=?"
	row := db.QueryRow(query, statistiqueID)

	// Récupère les données du client depuis la ligne de résultat
	var statistique Statistique
	err = row.Scan(&statistique.NbreChambreReservee, &statistique.NbreChambreLibre, &statistique.TauxOccupation, &statistique.NomHotel, &statistique.AdresseHotel)
	if err != nil {
		if err == sql.ErrNoRows {
			return Statistique{}, fmt.Errorf("room with ID %d not found", statistiqueID)
		}
		return Statistique{}, err
	}

	return statistique, nil
}
func getStatistiqueHandler(w http.ResponseWriter, r *http.Request) {
	// Récupère l'ID du client depuis l'URL
	statistiqueID, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/statistiques/"))
	if err != nil {
		http.Error(w, "Invalid  ID", http.StatusBadRequest)
		return
	}

	// Récupère le client correspondant depuis la base de données
	statistique, err := getStatistique(statistiqueID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Encode le client en JSON et renvoie la réponse HTTP
	json.NewEncoder(w).Encode(statistique)
}
func updateStatistiqueHandler(w http.ResponseWriter, r *http.Request) {
	// Récupère l'ID du client à mettre à jour depuis les paramètres de la requête
	statistiqueIDStr := r.URL.Query().Get("id")
	statistiqueID, err := strconv.Atoi(statistiqueIDStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	// Récupère les données du client à partir du corps de la requête
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	var updatedStatistique Statistique
	err = json.Unmarshal(reqBody, &updatedStatistique)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	// Ouvre une connexion à la base de données MySQL
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/hotel")
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Prépare et exécute la requête de mise à jour du client dans la base de données
	query := "UPDATE statistique  SET NbreChambreReservee=?, NbreChambreLibre=?,TauxOccupation=?,NomHotel=?,AdressHotel=?  WHERE id=?"
	_, err = db.Exec(query, updatedStatistique.NbreChambreReservee, updatedStatistique.NbreChambreLibre, updatedStatistique.TauxOccupation, updatedStatistique.NomHotel, updatedStatistique.AdresseHotel, statistiqueID)
	if err != nil {
		http.Error(w, "Failed to update statistics in database", http.StatusInternalServerError)
		return
	}

	// Retourne une réponse HTTP 200 OK
	w.WriteHeader(http.StatusOK)
}

func main() {
	//Client
	http.HandleFunc("/clients", getClientsHandler)
	http.HandleFunc("/clients/", getClientHandler)
	http.HandleFunc("/clients/new", createClientHandler)
	http.HandleFunc("/clients/update/", updateClientHandler)
	http.HandleFunc("/clients/delete/", deleteClientHandler)
	//Reservation
	http.HandleFunc("/reservations", getReservationsHandler)
	http.HandleFunc("/reservations/", getReservationHandler)
	http.HandleFunc("/reservations/update/", updateReservationHandler)
	http.HandleFunc("/reservations/new", createReservationHandler)
	http.HandleFunc("/reservations/delete/", deleteReservationHandler)
	//Chambre
	http.HandleFunc("/chambres", getChambresHandler)
	http.HandleFunc("/chambres/", getChambreHandler)
	http.HandleFunc("/chambres/new", createChambreHandler)
	http.HandleFunc("/chambres/update/", updateChambreHandler)
	http.HandleFunc("/chambres/delete/", deleteChambreHandler)
	//Statistique
	http.HandleFunc("/statistiques/", getStatistiqueHandler)
	http.HandleFunc("/statistiques/update/", updateStatistiqueHandler)

	// Lancement du serveur sur le port 8080
	log.Fatal(http.ListenAndServe(":9090", nil))

}
