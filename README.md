![alt text](https://raw.githubusercontent.com/LennardMa/mentalmaths/main/readme/Mental_Maths_Trainer.png)

The backend code for a mental maths trainer application written in Go using the Gin Framework.

This program does not use a database but stores necessary data in local memory.

To prevent problems and DOS conditions, I have implemented a custom middleware that limits the amount of API calls.

Additionally, a custom garbage collector implemented with cronjobs is cleaning up the local memory of unused items. 

# API Endpoints

![image](https://user-images.githubusercontent.com/74513606/166914307-4126dda0-5aaf-46f2-a8e0-84293b3b5207.png)

The /api endpoint takes JSON input with the syntax:
{"QuestionNumber": x, "MaxNumber": y}

QuestionNumber is the number of questions that should be calculated and send back to the client. 

MaxNumber is the highest real number that is part of the following questions.

The API call automatically calculates random questions depending on user input, sends them back to the client, calculates the correct answers and saves them together with an uuid in an array. 

![image](https://user-images.githubusercontent.com/74513606/166915191-83623b3a-8713-4607-b01a-71b03705061e.png)

When the client sends their answers back in the form of an array of ints, the function getScore compares the answers to the ones in local memory, computes a score based on accuracy, speed of completion and difficulty of questions. 

Afterwards the local memory entry gets delated and the session is complete. 

