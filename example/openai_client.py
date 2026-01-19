import time  # for measuring time duration of API calls
from openai import OpenAI
import os
def main():
    client = OpenAI(api_key="", base_url="http://localhost:8083/")

    start_time = time.time()

    # send a ChatCompletion request to count to 100
    response = client.chat.completions.create(
        model='gpt-4o',
        messages=[
            {'role': 'user', 'content': 'Count to 100, with a comma between each number and no newlines. E.g., 1, 2, 3, ...'}
        ],
        temperature=0,
    )
    # calculate the time it took to receive the response
    response_time = time.time() - start_time

    # print the time delay and text received
    print(f"Full response received {response_time:.2f} seconds after request")
    print(f"Full response received:\n{response}")

    start_time = time.time()

    # send a ChatCompletion request to count to 100
    response = client.chat.completions.create(
        model='gpt-4o',
        messages=[
            {'role': 'user', 'content': 'Count to 100, with a comma between each number and no newlines. E.g., 1, 2, 3, ...'}
        ],
        temperature=0,
    )
    # calculate the time it took to receive the response
    response_time = time.time() - start_time

    # print the time delay and text received
    print(f"Full response received {response_time:.2f} seconds after request")
    print(f"Full response received:\n{response}")

if __name__ == '__main__':
   main()