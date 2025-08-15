from flask import Flask, request, jsonify
import json

app = Flask(__name__)


@app.route("/", methods=["GET", "POST", "PUT", "DELETE"])
def handler():
    return jsonify(
        {
            "message": "Hello from Python function!",
            "method": request.method,
            "path": request.path,
            "headers": dict(request.headers),
            "body": request.get_data().decode("utf-8") if request.get_data() else None,
        }
    )


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=8080)
