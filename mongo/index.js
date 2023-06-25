const express = require('express');
//const mongoose = require('mongoose');

const app = express()

// Define a route
app.get('/', async (req, res) => {
  console.log("sc_nodemongoapi get request")
    res.sendStatus(200)
})

// Start the server
const port = 3001;
app.listen(port, () => {
  console.log(`Server listening on port ${port}`);
});
