# LLM-Powered Contextual Data Augmentation Pipeline

This project implements a pipeline for contextual data augmentation using Large Language Models (LLMs). It focuses on enhancing free-response questions by converting them into multiple-choice format, leveraging the power of LLMs for intelligent and context-aware transformations. The pipeline utilizes Apache Beam for scalable and parallel data processing.

## Key Features

- LLM-Powered Augmentation: Uses Anthropic's Claude API to intelligently generate multiple-choice options for questions
- Contextual Understanding: Leverages LLM's ability to understand question context, including exam sections and labels
- Scalable Processing: Utilizes Apache Beam for efficient, parallel processing of large question sets
- Data Validation: Ensures input questions meet required schema before processing
- Firestore Integration: Stores augmented questions for easy retrieval and further use

## How It Works

1. Input Processing: Reads free-response questions from a JSON input file
2. Data Validation: Validates the structure and content of each question
3. LLM Augmentation: Sends each question to the Claude API, which generates:
   - Four contextually relevant multiple-choice options
   - The correct answer
   - An explanation (if requested)
4. Data Transformation: Converts the free-response questions into multiple-choice format
5. Data Storage: Writes the augmented questions to Firestore for persistence

## Prerequisites

- Go 1.15 or higher
- Google Cloud Project with Firestore enabled
- **A service account with the Cloud Datastore User role assigned.**
- Anthropic API key for access to Claude
- Apache Beam SDK for Go

## Environment Setup

Set the following environment variables:

- `PROJECT_ID`: Your Google Cloud Project ID
- `COLLECTION`: Firestore collection name for storing augmented questions
- `CREDENTIALS_PATH`: Path to your Google Cloud credentials file
- `CLAUDE_API_KEY`: Your Anthropic API key

## Installation

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/llm-data-augmentation-pipeline.git
   cd llm-data-augmentation-pipeline
   ```

2. Install dependencies:
   ```
   go mod tidy
   ```

## Usage

Run the augmentation pipeline with:

```
go run . --input=path/to/your/questions.json
```

## Project Structure

- `main.go`: Entry point, sets up and runs the augmentation pipeline
- `llm_client.go`: Implements the LLM client for question augmentation
- `firestore_client.go`: Handles writing augmented questions to Firestore
- `question_parser.go`: Defines data structures and parsing logic for questions
- `file_utils.go`: Utility functions for file operations

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License:

```
MIT License

Copyright (c) [2024] [Fermin Blanco]

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```