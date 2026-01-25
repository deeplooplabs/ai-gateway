"""
LlamaIndex OpenAIEmbedding Example with AI Gateway

This example demonstrates how to use LlamaIndex's OpenAIEmbedding
with the AI Gateway to get text embeddings.

Prerequisites:
    pip install llama-index-embeddings-openai

Set environment variables:
    export AI_GATEWAY_URL="http://localhost:8083"
    export AI_GATEWAY_API_KEY="your-api-key"
"""

import os
from llama_index.embeddings.openai import OpenAIEmbedding

# Configuration
AI_GATEWAY_URL = os.getenv("AI_GATEWAY_URL", "http://localhost:8083")
AI_GATEWAY_API_KEY = os.getenv("AI_GATEWAY_API_KEY", "test-key")

# The api_base should point to the gateway without the /v1 suffix
# The gateway will handle the /v1/embeddings endpoint
API_BASE = f"{AI_GATEWAY_URL}/v1"


def example_single_text():
    """Get embedding for a single text."""
    print("\n=== Single Text Embedding ===")

    embed_model = OpenAIEmbedding(
        model="text-embedding-3-small",
        api_base=API_BASE,
        api_key=AI_GATEWAY_API_KEY,
    )

    text = "Hello, world!"
    embedding = embed_model.get_text_embedding(text)

    print(f"Text: {text}")
    print(f"Embedding dimension: {len(embedding)}")
    print(f"First 5 values: {embedding[:5]}")


def example_multiple_texts():
    """Get embeddings for multiple texts."""
    print("\n=== Multiple Texts Embedding ===")

    embed_model = OpenAIEmbedding(
        model="text-embedding-3-small",
        api_base=API_BASE,
        api_key=AI_GATEWAY_API_KEY,
    )

    texts = [
        "The quick brown fox jumps over the lazy dog.",
        "Machine learning is a subset of artificial intelligence.",
        "Python is a popular programming language.",
    ]

    embeddings = embed_model.get_text_embeddings(texts)

    print(f"Number of texts: {len(texts)}")
    print(f"Number of embeddings: {len(embeddings)}")
    for i, embedding in enumerate(embeddings):
        print(f"  Text {i+1}: dimension={len(embedding)}, first_3={embedding[:3]}")


def example_with_dimensions():
    """Get embedding with custom dimensions (v3 models only)."""
    print("\n=== Embedding with Custom Dimensions ===")

    embed_model = OpenAIEmbedding(
        model="text-embedding-3-small",
        api_base=API_BASE,
        api_key=AI_GATEWAY_API_KEY,
        dimensions=512,  # Custom embedding dimension
    )

    text = "Dimensionality reduction for embeddings."
    embedding = embed_model.get_text_embedding(text)

    print(f"Text: {text}")
    print(f"Requested dimensions: 512")
    print(f"Actual dimension: {len(embedding)}")


def example_query_vs_text():
    """Show difference between query and text embedding modes."""
    print("\n=== Query vs Text Embedding Modes ===")

    # TEXT_SEARCH_MODE is the default
    embed_model = OpenAIEmbedding(
        mode="text_search",
        model="text-embedding-3-small",
        api_base=API_BASE,
        api_key=AI_GATEWAY_API_KEY,
    )

    query = "What is machine learning?"
    document = "Machine learning is a branch of artificial intelligence."

    query_embedding = embed_model.get_query_embedding(query)
    text_embedding = embed_model.get_text_embedding(document)

    print(f"Query: {query}")
    print(f"Query embedding dimension: {len(query_embedding)}")
    print(f"Document: {document}")
    print(f"Text embedding dimension: {len(text_embedding)}")


def example_with_large_batch():
    """Handle large batch of texts with automatic batching."""
    print("\n=== Large Batch Embedding ===")

    embed_model = OpenAIEmbedding(
        model="text-embedding-3-small",
        api_base=API_BASE,
        api_key=AI_GATEWAY_API_KEY,
        embed_batch_size=10,  # Process 10 texts at a time
    )

    # Generate sample texts
    texts = [f"This is sample text number {i}." for i in range(25)]

    embeddings = embed_model.get_text_embeddings_batch(texts)

    print(f"Total texts: {len(texts)}")
    print(f"Batch size: 10")
    print(f"Total embeddings received: {len(embeddings)}")
    print(f"Each embedding dimension: {len(embeddings[0]) if embeddings else 0}")


def main():
    """Run all examples."""
    print("=" * 60)
    print("LlamaIndex OpenAIEmbedding with AI Gateway Examples")
    print("=" * 60)
    print(f"\nGateway URL: {AI_GATEWAY_URL}")
    print(f"API Base: {API_BASE}")

    try:
        example_single_text()
        example_multiple_texts()
        example_with_dimensions()
        example_query_vs_text()
        example_with_large_batch()

        print("\n" + "=" * 60)
        print("All examples completed successfully!")
        print("=" * 60)

    except Exception as e:
        print(f"\nError: {e}")
        print("\nTroubleshooting:")
        print("1. Ensure AI Gateway is running:")
        print(f"   cd example && go run main.go")
        print("2. Check the gateway URL is correct:")
        print(f"   Current: {AI_GATEWAY_URL}")
        print("3. Verify the model is registered in the gateway")
        print("4. Check gateway logs for detailed error messages")


if __name__ == "__main__":
    main()
