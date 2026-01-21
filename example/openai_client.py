import json  # for JSON parsing
import time  # for measuring time duration of API calls
from openai import OpenAI
import os
import requests

def main():
    client = OpenAI(api_key="1234566:team-1", base_url="http://localhost:8083/v1")

    print("=" * 60)
    print("Test 1: OpenAI ChatCompletion API")
    print("=" * 60)

    start_time = time.time()

    # send a ChatCompletion request to count to 100
    response = client.chat.completions.create(
        model='gpt-4o',
        messages=[
            {'role': 'user', 'content': 'Count to 10, with a comma between each number and no newlines. E.g., 1, 2, 3, ...'}
        ],
        temperature=0,
    )
    # calculate the time it took to receive the response
    response_time = time.time() - start_time

    # print the time delay and text received
    print(f"Full response received {response_time:.2f} seconds after request")
    print(f"Full response received:\n{response}\n")

    print("=" * 60)
    print("Test 2: OpenResponses API")
    print("=" * 60)

    start_time = time.time()

    # Send an OpenResponses request
    or_response = requests.post(
        "http://localhost:8083/v1/responses",
        headers={
            "Authorization": "Bearer 1234566:team-1",
            "Content-Type": "application/json",
        },
        json={
            "model": "gpt-4o",
            "input": "Count to 10, with a comma between each number and no newlines. E.g., 1, 2, 3, ...",
        },
    )
    response_time = time.time() - start_time

    print(f"Full response received {response_time:.2f} seconds after request")
    print(f"Status Code: {or_response.status_code}")
    if or_response.status_code == 200:
        data = or_response.json()
        print(f"Response ID: {data.get('id')}")
        print(f"Status: {data.get('status')}")
        print(f"Model: {data.get('model')}")
        print(f"Output:")
        for item in data.get('output', []):
            if item.get('type') == 'message':
                for content in item.get('content', []):
                    if content.get('type') == 'output_text':
                        print(f"  {content.get('text')}")
        if data.get('usage'):
            print(f"Usage: {data['usage']}")
    else:
        print(f"Error: {or_response.text}")

    print("\n" + "=" * 60)
    print("Test 3: OpenResponses Streaming API")
    print("=" * 60)

    start_time = time.time()

    # Send an OpenResponses streaming request
    or_response = requests.post(
        "http://localhost:8083/v1/responses",
        headers={
            "Authorization": "Bearer 1234566:team-1",
            "Content-Type": "application/json",
        },
        json={
            "model": "gpt-4o",
            "input": "Count to 5, with a comma between each number.",
            "stream": True,
        },
        stream=True,
    )

    print(f"Streaming events:")
    for line in or_response.iter_lines():
        if line:
            line_str = line.decode('utf-8')
            if line_str.startswith('data: '):
                data_str = line_str[6:]  # Remove 'data: ' prefix
                if data_str == '[DONE]':
                    print(f"  [DONE]")
                    break
                # Parse the event data
                try:
                    event = json.loads(data_str)
                    event_type = event.get('type', 'unknown')
                    print(f"  Event: {event_type}")
                except json.JSONDecodeError:
                    print(f"  Raw: {data_str}")

    response_time = time.time() - start_time
    print(f"Streaming completed in {response_time:.2f} seconds\n")

if __name__ == '__main__':
   main()