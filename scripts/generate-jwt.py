#!/usr/bin/env python3
"""
JWT Token Generator for Generic Rule Engine Testing

This script generates JWT tokens for testing the rule engine APIs.
It can be used as a standalone tool or imported by other scripts.
"""

import argparse
import datetime
import json
import sys
from typing import Dict, Any

# Try to import PyJWT with better error handling
try:
    import jwt
except ImportError:
    print("Error: PyJWT library is required.")
    print("Install it with one of these commands:")
    print("  pip install PyJWT")
    print("  pip3 install PyJWT")
    print("  conda install pyjwt")
    print("If you're using a virtual environment, make sure it's activated.")
    sys.exit(1)


def generate_jwt_token(
    client_id: str,
    role: str,
    secret: str,
    expiry_hours: int = 24,
    algorithm: str = "HS256"
) -> str:
    """
    Generate a JWT token with the specified claims.
    
    Args:
        client_id: The client identifier
        role: The user role (admin, viewer, executor)
        secret: The JWT secret key
        expiry_hours: Token expiration time in hours
        algorithm: JWT signing algorithm
        
    Returns:
        The encoded JWT token string
    """
    now = datetime.datetime.utcnow()
    expiry = now + datetime.timedelta(hours=expiry_hours)
    
    claims = {
        "clientId": client_id,
        "role": role,
        "exp": expiry,
        "iat": now,
        "nbf": now
    }
    
    try:
        token = jwt.encode(claims, secret, algorithm=algorithm)
        return token
    except Exception as e:
        raise ValueError(f"Failed to generate JWT token: {e}")


def decode_jwt_token(token: str, secret: str) -> Dict[str, Any]:
    """
    Decode and verify a JWT token.
    
    Args:
        token: The JWT token string
        secret: The JWT secret key
        
    Returns:
        The decoded claims dictionary
    """
    try:
        claims = jwt.decode(token, secret, algorithms=["HS256"])
        return claims
    except jwt.ExpiredSignatureError:
        raise ValueError("Token has expired")
    except jwt.InvalidTokenError as e:
        raise ValueError(f"Invalid token: {e}")


def main():
    """Main function for command-line usage."""
    parser = argparse.ArgumentParser(
        description="Generate JWT tokens for Generic Rule Engine testing"
    )
    parser.add_argument(
        "--client-id",
        default="test-client",
        help="Client ID (default: test-client)"
    )
    parser.add_argument(
        "--role",
        default="admin",
        choices=["admin", "viewer", "executor"],
        help="User role (default: admin)"
    )
    parser.add_argument(
        "--secret",
        default="dev-secret-key-change-in-production",
        help="JWT secret key"
    )
    parser.add_argument(
        "--expiry",
        type=int,
        default=24,
        help="Token expiration time in hours (default: 24)"
    )
    parser.add_argument(
        "--decode",
        help="Decode and verify an existing token"
    )
    parser.add_argument(
        "--format",
        choices=["token", "json", "curl"],
        default="token",
        help="Output format (default: token)"
    )
    
    args = parser.parse_args()
    
    if args.decode:
        # Decode existing token
        try:
            claims = decode_jwt_token(args.decode, args.secret)
            if args.format == "json":
                print(json.dumps(claims, default=str, indent=2))
            else:
                print(f"Token is valid. Claims: {claims}")
        except ValueError as e:
            print(f"Error: {e}")
            sys.exit(1)
    else:
        # Generate new token
        try:
            token = generate_jwt_token(
                args.client_id,
                args.role,
                args.secret,
                args.expiry
            )
            
            if args.format == "json":
                claims = decode_jwt_token(token, args.secret)
                output = {
                    "token": token,
                    "claims": claims,
                    "usage": f"curl -H \"Authorization: Bearer {token}\" http://localhost:8080/v1/namespaces"
                }
                print(json.dumps(output, default=str, indent=2))
            elif args.format == "curl":
                print(f"curl -H \"Authorization: Bearer {token}\" http://localhost:8080/v1/namespaces")
            else:
                print(token)
                
        except ValueError as e:
            print(f"Error: {e}")
            sys.exit(1)


if __name__ == "__main__":
    main() 