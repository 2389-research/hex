#!/usr/bin/env python3
# /// script
# requires-python = ">=3.12"
# dependencies = [
#     "requests>=2.32.5",
#     "python-dotenv>=1.0.0",
# ]
# ///
"""
Fetch current model lists from OpenAI, Gemini, and OpenRouter APIs.
This script queries the actual API endpoints to get real, current model names.
"""

import os
import sys
import json
import requests
from pathlib import Path
from typing import Dict, List, Optional
from dotenv import load_dotenv

# Load .env file from project root
env_path = Path(__file__).parent.parent / '.env'
if env_path.exists():
    load_dotenv(env_path)
    print(f"✅ Loaded environment from {env_path}", file=sys.stderr)


def fetch_openai_models(api_key: Optional[str] = None) -> List[str]:
    """Fetch available models from OpenAI API."""
    api_key = api_key or os.getenv("OPENAI_API_KEY")
    if not api_key:
        print("⚠️  OPENAI_API_KEY not set, skipping OpenAI", file=sys.stderr)
        return []

    try:
        response = requests.get(
            "https://api.openai.com/v1/models",
            headers={"Authorization": f"Bearer {api_key}"},
            timeout=10
        )
        response.raise_for_status()
        data = response.json()

        # Filter for GPT models (not embeddings, whisper, etc)
        models = [
            model["id"]
            for model in data.get("data", [])
            if model["id"].startswith(("gpt-", "chatgpt-", "o1-", "o3-"))
        ]
        return sorted(models)
    except Exception as e:
        print(f"❌ Error fetching OpenAI models: {e}", file=sys.stderr)
        return []


def fetch_gemini_models(api_key: Optional[str] = None) -> List[str]:
    """Fetch available models from Google Gemini API."""
    api_key = api_key or os.getenv("GEMINI_API_KEY") or os.getenv("GOOGLE_API_KEY")
    if not api_key:
        print("⚠️  GEMINI_API_KEY not set, skipping Gemini", file=sys.stderr)
        return []

    try:
        response = requests.get(
            f"https://generativelanguage.googleapis.com/v1beta/models?key={api_key}",
            timeout=10
        )
        response.raise_for_status()
        data = response.json()

        # Extract model names that support generateContent
        models = []
        for model in data.get("models", []):
            name = model.get("name", "").replace("models/", "")
            # Only include models that support text generation
            methods = model.get("supportedGenerationMethods", [])
            if "generateContent" in methods and name.startswith("gemini-"):
                models.append(name)

        return sorted(models)
    except Exception as e:
        print(f"❌ Error fetching Gemini models: {e}", file=sys.stderr)
        return []


def fetch_openrouter_models(api_key: Optional[str] = None) -> List[str]:
    """Fetch available models from OpenRouter API."""
    api_key = api_key or os.getenv("OPENROUTER_API_KEY")
    if not api_key:
        print("⚠️  OPENROUTER_API_KEY not set, skipping OpenRouter", file=sys.stderr)
        return []

    try:
        response = requests.get(
            "https://openrouter.ai/api/v1/models",
            headers={"Authorization": f"Bearer {api_key}"},
            timeout=10
        )
        response.raise_for_status()
        data = response.json()

        # Get model IDs - OpenRouter uses provider/model format
        models = [model["id"] for model in data.get("data", [])]

        # Filter for major providers (Anthropic, OpenAI, Google)
        filtered = [
            m for m in models
            if any(m.startswith(p) for p in ["anthropic/", "openai/", "google/"])
        ]

        return sorted(filtered)
    except Exception as e:
        print(f"❌ Error fetching OpenRouter models: {e}", file=sys.stderr)
        return []


def main():
    """Fetch and display current models from all providers."""
    print("🔍 Fetching current models from API providers...\n")

    # Fetch from each provider
    openai_models = fetch_openai_models()
    gemini_models = fetch_gemini_models()
    openrouter_models = fetch_openrouter_models()

    # Display results
    print("=" * 80)
    print("OPENAI MODELS")
    print("=" * 80)
    if openai_models:
        # Show newest first (by name, assuming version numbers sort correctly)
        for model in reversed(openai_models[-10:]):  # Last 10 (newest)
            print(f"  {model}")
        print(f"\nTotal: {len(openai_models)} models")
    else:
        print("  (No models fetched)")

    print("\n" + "=" * 80)
    print("GEMINI MODELS")
    print("=" * 80)
    if gemini_models:
        for model in gemini_models:
            print(f"  {model}")
        print(f"\nTotal: {len(gemini_models)} models")
    else:
        print("  (No models fetched)")

    print("\n" + "=" * 80)
    print("OPENROUTER MODELS (Anthropic, OpenAI, Google)")
    print("=" * 80)
    if openrouter_models:
        # Group by provider
        anthropic = [m for m in openrouter_models if m.startswith("anthropic/")]
        openai = [m for m in openrouter_models if m.startswith("openai/")]
        google = [m for m in openrouter_models if m.startswith("google/")]

        print("\nAnthropic:")
        for model in anthropic[:5]:  # First 5
            print(f"  {model}")

        print("\nOpenAI:")
        for model in openai[:5]:  # First 5
            print(f"  {model}")

        print("\nGoogle:")
        for model in google[:5]:  # First 5
            print(f"  {model}")

        print(f"\nTotal: {len(openrouter_models)} models")
    else:
        print("  (No models fetched)")

    # Save to JSON for programmatic use
    output = {
        "openai": openai_models,
        "gemini": gemini_models,
        "openrouter": {
            "anthropic": [m for m in openrouter_models if m.startswith("anthropic/")],
            "openai": [m for m in openrouter_models if m.startswith("openai/")],
            "google": [m for m in openrouter_models if m.startswith("google/")],
        }
    }

    output_file = "scripts/models.json"
    with open(output_file, "w") as f:
        json.dump(output, f, indent=2)

    print(f"\n💾 Saved to {output_file}")
    print("\n" + "=" * 80)
    print("RECOMMENDED MODELS FOR ERROR MESSAGES")
    print("=" * 80)

    # Suggest best models for error messages
    if openai_models:
        latest_gpt = [m for m in openai_models if m.startswith("gpt-")][-3:]
        print(f"\nOpenAI: {', '.join(latest_gpt)}")

    if gemini_models:
        stable = [m for m in gemini_models if "preview" not in m and "exp" not in m][-3:]
        print(f"Gemini: {', '.join(stable)}")

    if openrouter_models:
        anthropic_latest = [m for m in openrouter_models if m.startswith("anthropic/claude")][:3]
        print(f"OpenRouter: {', '.join(anthropic_latest)}")


if __name__ == "__main__":
    main()
