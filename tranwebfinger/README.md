# Self-Evolving Wappalyzer Rules System

A powerful, self-evolving website technology detection system based on Wappalyzer rules format.

## Features

- **Technology Detection**: Detects web technologies using headers, HTML patterns, and script patterns
- **Self-Evolving**: Automatically analyzes websites to discover and add new detection rules
- **Customizable**: Easily add, remove, or modify technology rules
- **Extensible**: Compatible with Wappalyzer's technologies.json format
- **Statistics**: Provides detailed statistics about detected technologies and evolution events

## Installation

```bash
# Install dependencies
pip3 install requests

# Run the system
python3 main.py
```

## Files

- `main.py`: Core Python script with the SelfEvolvingWappalyzer class
- `technologies.json`: Technology detection rules (Wappalyzer compatible)
- `config.json`: Configuration file for the system
- `README.md`: This documentation file

## Usage

### Basic Usage

```python
from main import SelfEvolvingWappalyzer

# Create instance
wappalyzer = SelfEvolvingWappalyzer()

# Show statistics
wappalyzer.show_stats()
```

### Scan a Website

```python
# Scan a website for technologies
detected = wappalyzer.scan("https://example.com")
```

### Evolve Rules

```python
# Define expected technologies
expected = {
    "ExampleTech": {
        "name": "Example Technology",
        "category": "Web Servers"
    }
}

# Evolve rules by analyzing a website
wappalyzer.evolve("https://example.com", expected)
```

### Add Technology Manually

```python
wappalyzer.add_technology(
    "ExampleJS",
    "Example JavaScript Library",
    "JavaScript Libraries",
    "An example JavaScript library",
    "https://example.com",
    {
        "scripts": ["example\\.js", "example-lib\\.js"]
    }
)
```

### Remove Technology

```python
wappalyzer.remove_technology("ExampleJS")
```

## Configuration

Edit `config.json` to configure the system:

```json
{
  "self_evolving_wappalyzer": {
    "rules_file": "technologies.json",
    "timeout": 10,
    "scan_headers": true,
    "scan_html": true,
    "scan_scripts": true,
    "evolution_enabled": true,
    "confidence_threshold": 0.7,
    "log_evolution": true,
    "max_evolution_events": 100
  }
}
```

## Rule Format

Each technology in `technologies.json` has the following structure:

```json
"TechnologyName": {
  "name": "Display Name",
  "category": "Category Name",
  "description": "Description of the technology",
  "website": "https://example.com",
  "headers": {
    "Header-Name": ["pattern1", "pattern2"]
  },
  "html": ["<meta name=\"generator\" content=\"Pattern"],
  "scripts": ["script-name\\.js", "library-name"]
}
```

## Categories

Categories help organize technologies. Each category has a name and priority:

```json
"CategoryName": {
  "name": "Display Name",
  "priority": 1
}
```

## Evolution Mechanism

The system evolves by:

1. Scanning a website for known technologies
2. Identifying technologies that should be detected but aren't
3. Analyzing the website's headers, HTML, and scripts for patterns
4. Adding new detection rules for missing technologies
5. Saving the updated rules to the JSON file

## Example Output

```
=== Self-Evolving Wappalyzer Statistics ===
Total technologies: 3
Total categories: 3
Evolution events: 0

Technologies by category:
- CMS: 1
- JavaScript Frameworks: 1
- Web Servers: 1
```

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
