FROM ollama/ollama:latest

# Start the Ollama service in the background
RUN ollama serve & \
    # Give the service a few seconds to start up (adjust if needed)
    sleep 10 && \
    # Pull the llama3.2 model
    ollama pull llama3.2 && \
    # Terminate the background service so the build can complete
    pkill -f "ollama serve"
