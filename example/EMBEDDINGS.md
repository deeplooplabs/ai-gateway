# LlamaIndex Embeddings Example

This example demonstrates how to use [LlamaIndex](https://llamaindex.ai/) with the AI Gateway for text embeddings.

## Prerequisites

### 1. Start AI Gateway

```bash
cd example
go run main.go
```

The gateway will start on `http://localhost:8083`.

### 2. Install Python Dependencies

```bash
pip install llama-index-embeddings-openai
```

## Quick Start

### Basic Usage

```python
from llama_index.embeddings.openai import OpenAIEmbedding

# Configure embedding model to use AI Gateway
embed_model = OpenAIEmbedding(
    model="text-embedding-3-small",
    api_base="http://localhost:8083/v1",
    api_key="your-api-key",
)

# Get embedding for a single text
embedding = embed_model.get_text_embedding("Hello, world!")
print(f"Dimension: {len(embedding)}")
```

### With LlamaIndex Indices

```python
from llama_index.core import VectorStoreIndex, SimpleDirectoryReader
from llama_index.embeddings.openai import OpenAIEmbedding
from llama_index.llms.openai import OpenAI

# Configure embeddings and LLM to use AI Gateway
embed_model = OpenAIEmbedding(
    model="text-embedding-3-small",
    api_base="http://localhost:8083/v1",
    api_key="your-api-key",
)

llm = OpenAI(
    model="gpt-4o",
    api_base="http://localhost:8083/v1",
    api_key="your-api-key",
)

# Load documents and create index
documents = SimpleDirectoryReader("data").load_data()
index = VectorStoreIndex.from_documents(documents, embed_model=embed_model)

# Query with configured LLM
query_engine = index.as_query_engine(llm=llm)
response = query_engine.query("What is the document about?")
print(response)
```

## Available Models

Based on the gateway configuration, the following models are available:

| Model Name | Rewritten To | Type |
|------------|--------------|------|
| `text-embedding-3-small` | `Qwen/Qwen3-Embedding-4B` | embedding |
| `text-embedding-3-large` | `Qwen/Qwen3-Embedding-8B` | embedding |
| `text-embedding-ada-002` | `Qwen/Qwen3-Embedding-8B` | embedding |

## Configuration Options

```python
embed_model = OpenAIEmbedding(
    model="text-embedding-3-small",
    api_base="http://localhost:8083/v1",  # Gateway URL
    api_key="your-api-key",                # Authentication

    # Optional parameters
    dimensions=512,           # Custom dimensions (v3 models only)
    embed_batch_size=100,     # Batch size for multiple texts
    max_retries=10,           # Retry attempts
    timeout=60.0,             # Request timeout in seconds
)
```

## Running the Example

```bash
# Set environment variables (optional)
export AI_GATEWAY_URL="http://localhost:8083"
export AI_GATEWAY_API_KEY="your-api-key"

# Run the example
python example/embeddings_example.py
```

## Expected Output

```
============================================================
LlamaIndex OpenAIEmbedding with AI Gateway Examples
============================================================

Gateway URL: http://localhost:8083
API Base: http://localhost:8083/v1

=== Single Text Embedding ===
Text: Hello, world!
Embedding dimension: 1024
First 5 values: [0.0123, -0.0456, 0.0789, ...]

=== Multiple Texts Embedding ===
Number of texts: 3
Number of embeddings: 3
  Text 1: dimension=1024, first_3=[...]
  Text 2: dimension=1024, first_3=[...]
  Text 3: dimension=1024, first_3=[...]

============================================================
All examples completed successfully!
============================================================
```

## Troubleshooting

### Connection Error

```
Error: Connection refused
```

**Solution**: Ensure the AI Gateway is running:
```bash
cd example && go run main.go
```

### Model Not Found

```
Error code: 404 - {'error': {'message': 'model not found: xxx'}}
```

**Solution**: Verify the model is registered in `example/main.go`:
```go
registry.RegisterWithOptions("text-embedding-3-small", openAIProvider,
    model.WithPreferredAPI(provider.APITypeEmbeddings),
)
```

### Provider Request Failed

```
Error code: 502 - {'error': {'message': 'provider request failed'}}
```

**Solution**: Check the upstream provider configuration:
- Verify `OPENAI_BASE_URL` is correct
- Verify `OPENAI_API_KEY` is valid
- Check gateway logs for detailed error messages

## Gateway Logs

When embeddings requests are processed, you'll see logs like:

```
INFO Registered model model=text-embedding-3-small type=embedding provider=http rewrite_to=Qwen/Qwen3-Embedding-4B
INFO Embeddings request received model=text-embedding-3-small encoding_format= dimensions=0
INFO Model rewrite applied original=text-embedding-3-small rewritten=Qwen/Qwen3-Embedding-4B
INFO Provider resolved provider=http supported_apis=all
INFO Embeddings response successful embedding_count=1 model=Qwen/Qwen3-Embedding-4B prompt_tokens=5
```
