# Ask Chotu

**Ask Chotu** is a simple chatbot project that allows you to ask questions about yourself or any other topic. It leverages open-source large language models (LLMs) to generate conversational responses.
<img width="404" height="316" alt="image" src="https://github.com/user-attachments/assets/ad317e8d-39b1-41df-b7d5-8dc8248ffbd5" />


## Features

- Ask questions and get intelligent, context-aware answers.
- Easily switch between different LLM backends (e.g., NLPCloud, Ollama, Hugging Face, etc.).
- Simple Go-based backend for easy integration and extension.

## Getting Started

### Prerequisites

- [Go](https://golang.org/) 1.18 or higher
- (Optional) An LLM backend such as [Ollama](https://github.com/jmorganca/ollama) running locally or an API key for a hosted model

### Installation

1. Clone the repository:
    ```sh
    git clone https://github.com/Allen-Career-Institute/ask-chotu.git
    cd ask-chotu
    ```

2. Install dependencies:
    ```sh
    go mod tidy
    ```

3. Configure your preferred LLM backend in `ai.go`.

### Running

To start the chatbot server:

```sh
go run main.go
```

## Usage

- Interact with Chotu via the provided interface (CLI, API, or web—depending on your implementation).
- Ask questions about yourself or any topic, and Chotu will respond using the configured language model.

## Customization

- You can switch the LLM backend by modifying the `getChatbotResponse` function in `ai.go`.
- Add new features or integrations as needed.

## License

This project is licensed under the MIT License.

---

**Ask Chotu** – Your friendly, personal AI assistant!
