"""Subprocess wrapper for openSeaChest CLI tools."""

from __future__ import annotations

import logging
import subprocess
from typing import Optional

logger = logging.getLogger(__name__)

DEFAULT_TIMEOUT = 120  # seconds


class CLIError(Exception):
    """Raised when an openSeaChest CLI command fails."""

    def __init__(self, command: list[str], returncode: int, stderr: str):
        self.command = command
        self.returncode = returncode
        self.stderr = stderr
        super().__init__(
            f"Command {' '.join(command)} exited with code {returncode}: "
            f"{stderr[:500]}"
        )


class CLIRunner:
    """Wrapper around openSeaChest CLI executables.

    Args:
        basics_path: Path to the openSeaChest_Basics executable.
        smart_path: Path to the openSeaChest_SMART executable.
        timeout: Timeout in seconds for each CLI invocation.
    """

    def __init__(
        self,
        basics_path: str = "openSeaChest_Basics",
        smart_path: str = "openSeaChest_SMART",
        timeout: int = DEFAULT_TIMEOUT,
    ):
        self.basics_path = basics_path
        self.smart_path = smart_path
        self.timeout = timeout

    def _run(self, command: list[str]) -> str:
        """Execute a command and return its stdout.

        Args:
            command: Full command as a list of strings.

        Returns:
            The stdout output as a string.

        Raises:
            CLIError: If the command exits with a non-zero return code.
        """
        logger.debug("Running: %s", " ".join(command))
        try:
            result = subprocess.run(
                command,
                capture_output=True,
                text=True,
                timeout=self.timeout,
            )
        except subprocess.TimeoutExpired as exc:
            raise CLIError(
                command, -1, f"Command timed out after {self.timeout}s"
            ) from exc
        except FileNotFoundError as exc:
            raise CLIError(
                command, -1, f"Executable not found: {command[0]}"
            ) from exc

        if result.returncode != 0:
            raise CLIError(command, result.returncode, result.stderr)

        return result.stdout

    def scan(self) -> str:
        """Run ``openSeaChest_Basics --scan`` and return stdout."""
        return self._run([self.basics_path, "--scan"])

    def device_info(self, device_path: str) -> str:
        """Run ``openSeaChest_Basics --deviceInfo`` for a specific device."""
        return self._run([
            self.basics_path, "-d", device_path, "--deviceInfo",
        ])

    def smart_check(self, device_path: str) -> str:
        """Run ``openSeaChest_SMART --smartCheck`` for a specific device."""
        return self._run([
            self.smart_path, "-d", device_path, "--smartCheck",
        ])

    def smart_attributes(self, device_path: str) -> str:
        """Run ``openSeaChest_SMART --smartAttributes raw`` for a device."""
        return self._run([
            self.smart_path, "-d", device_path, "--smartAttributes", "raw",
        ])
