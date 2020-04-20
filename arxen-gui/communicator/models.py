from django.db import models

# Create your models here.


class Chat(models.Model):
    uuid = models.CharField(max_length=100)


class Message(models.Model):
    message_text = models.CharField(max_length=1000)
    pub_date = models.DateTimeField('date published')
    author = models.TextField(max_length=100)
    chat = models.ForeignKey(Chat, on_delete=models.CASCADE)
