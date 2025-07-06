const express = require('express')
const axios = require('axios')
const app = express()

app.use(express.json({ limit: '2mb' }))
app.use(require('cors')())

app.post('/proxy', async (req, res) => {
	const { url, method = 'POST', headers = {}, data } = req.body

	try {
		const response = await axios({
			url,
			method,
			headers,
			data,
			validateStatus: () => true,
		})

		res.status(response.status).json({
			status: response.status,
			headers: response.headers,
			data: response.data,
		})
	} catch (err) {
		res.status(500).json({ error: err.message })
	}
})

const PORT = 8085
app.listen(PORT, () => {
	console.log(`Postman bridge running on http://localhost:${PORT}`)
})
