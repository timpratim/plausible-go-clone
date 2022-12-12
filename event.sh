curl http://localhost:8080/events \
    --include \
    --header "Content-Type: application/json" \
    --request "POST" \
    --data '{"name": "Signup"}'