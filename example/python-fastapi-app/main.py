from fastapi import FastAPI
from pydantic import BaseModel

app = FastAPI(
    title="Greeting API",
    description="A simple API to greet users by name.",
    version="1.0.0",
)


class Greeting(BaseModel):
    message: str


@app.get(
    "/greet/{name}",
    summary="Greet a user by their name",
    description="This endpoint accepts a name as a path parameter and returns a personalized greeting message.",
    tags=["Greetings"],
    response_model=Greeting,
)
async def greet_user(name: str):
    greeting_message = f"Hello, {name}! Welcome to the API."
    return Greeting(message=greeting_message)
