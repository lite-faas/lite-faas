const express = require("express");
const app = express();

app.use(express.json());
app.use(express.text());

app.all("/", (req, res) => {
    res.json({
        message: "Hello from Node.js function!",
        method: req.method,
        path: req.path,
        headers: req.headers,
        body: req.body,
    });
});

app.listen(8080, "0.0.0.0", () => {
    console.log("Node.js function server running on port 8080");
});
