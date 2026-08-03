import os
import re
from pathlib import Path
from typing import Any

import yaml

_ENV_PATTERN = re.compile(r"^\$\{([A-Za-z_][A-Za-z0-9_]*)\}$")


def _substitute(value: Any) -> Any:
    if isinstance(value, str):
        match = _ENV_PATTERN.match(value)
        if match:
            env_name = match.group(1)
            env_value = os.getenv(env_name)
            if env_value is None or env_value == "":
                raise RuntimeError(f"Mandatory environment variable {env_name} not found")
            return env_value
        return value
    if isinstance(value, list):
        return [_substitute(item) for item in value]
    if isinstance(value, dict):
        return {key: _substitute(item) for key, item in value.items()}
    return value


def load_config(path: Path | None = None) -> dict[str, Any]:
    config_path = path or Path("config.yaml")
    if not config_path.is_file():
        raise FileNotFoundError(f"config file not found: {config_path}")

    raw = yaml.safe_load(config_path.read_text())
    if not isinstance(raw, dict):
        raise RuntimeError(f"invalid config file: {config_path}")
    return _substitute(raw)
