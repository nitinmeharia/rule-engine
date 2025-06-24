#!/usr/bin/env python3
"""
Simple JWT Token Generator for Testing
This script generates JWT tokens for API testing with admin role.
"""

import jwt
import datetime
import argparse
import sys

def generate_jwt_token(secret="dev-secret-key-change-in-production", client_id="test-client", role="admin", expires_in_hours=24):
    """
    Generate a JWT token for testing purposes.
    
    Args:
        secret (str): JWT secret key
        client_id (str): Client identifier
        role (str): User role (admin, viewer, executor)
        expires_in_hours (int): Token expiration time in hours
    
    Returns:
        str: Generated JWT token
    """
    
    # Token payload
    payload = {
        'clientId': client_id,
        'role': role,
        'exp': datetime.datetime.now(datetime.timezone.utc) + datetime.timedelta(hours=expires_in_hours),
        'iat': datetime.datetime.now(datetime.timezone.utc),
        'iss': 'rule-engine-test'
    }
    
    # Generate token
    token = jwt.encode(payload, secret, algorithm='HS256')
    return token

def main():
    parser = argparse.ArgumentParser(description='Generate JWT token for API testing')
    parser.add_argument('--secret', default='dev-secret-key-change-in-production', help='JWT secret key')
    parser.add_argument('--client-id', default='test-client', help='Client identifier')
    parser.add_argument('--role', default='admin', choices=['admin', 'viewer', 'executor'], help='User role')
    parser.add_argument('--expires', type=int, default=24, help='Token expiration time in hours')
    parser.add_argument('--export', action='store_true', help='Export as environment variable')
    parser.add_argument('--quiet', action='store_true', help='Output only the token (for scripting)')
    
    args = parser.parse_args()
    
    try:
        token = generate_jwt_token(
            secret=args.secret,
            client_id=args.client_id,
            role=args.role,
            expires_in_hours=args.expires
        )
        
        if args.export:
            print(f"export JWT_TOKEN='{token}'")
        elif args.quiet:
            print(token)
        else:
            print("Generated JWT Token:")
            print(f"{token}")
            print("\nTo use in tests:")
            print(f"export JWT_TOKEN='{token}'")
            
    except Exception as e:
        print(f"Error generating token: {e}", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main() 