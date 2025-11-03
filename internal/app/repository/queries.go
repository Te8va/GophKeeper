package repository

const (

	// Authorization queries
	queryInsertUser = `
		INSERT INTO users(user_id, password, token)
		VALUES($1, $2, $3)
	`

	queryGetUserByLogin = `
		SELECT id, user_id, password, token
		FROM users
		WHERE user_id = $1
	`
	queryGetUserByID = `
		SELECT id, user_id, password, token
		FROM users
		WHERE id = $1
	`
	//Data
	saveDataItemQuery = `
	INSERT INTO data_items (user_id, type, data, metadata)
	VALUES ($1, $2, $3, $4)
	RETURNING id
	`

	getDataByIDQuery = `
		SELECT id, user_id, type, data, metadata
		FROM data_items 
		WHERE id = $1 AND user_id = $2
	`
	getAllUserDataQuery = `
		SELECT id, user_id, type, data, metadata
		FROM data_items 
		WHERE user_id = $1
		ORDER BY id DESC
	`
	updateDataQuery = `
		UPDATE data_items 
		SET type = $1, data = $2, metadata = $3
		WHERE id = $4 AND user_id = $5
	`
	deleteDataQuery = `
		DELETE FROM data_items 
		WHERE id = $1 AND user_id = $2
	`
)
