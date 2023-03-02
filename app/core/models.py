from django.db import models
from django.contrib.auth.models import AbstractBaseUser, PermissionsMixin, BaseUserManager
from django.conf import settings

def upload_location(instance, filename):
	file_path = 'profile/{user_id}-{filename}.jpg'.format(
			user_id = str(instance.id),
            filename = str(filename)
		)
	return file_path

class UserManager(BaseUserManager):
    def create_user(self, email, username, password=None, **extra_fields):
        if not email:
            raise ValueError("Email shouldn't be empty.")

        email = self.normalize_email(email)
        user = self.model(email=email, username=username, **extra_fields)

        user.set_password(password)
        user.save(using=self._db)

        return user

    def create_superuser(self, email, username, password):
        user = self.create_user(email, username, password)
        user.is_superuser = True
        user.is_staff = True
        user.save(using=self._db)

        return user


class User(AbstractBaseUser, PermissionsMixin):
    class UserType(models.IntegerChoices):
        Free = 1
        Premium = 2

    class OAuthType(models.IntegerChoices):
        Google = 1

    id = models.AutoField(primary_key=True)
    email = models.EmailField(max_length=255, unique=True)
    username = models.CharField(max_length=255, unique=True)
    image = models.ImageField(null=True, upload_to=upload_location)
    fcm_token = models.CharField(max_length=255, default="")
    refresh_token = models.CharField(max_length=255, null=True, default=None)
    user_type = models.IntegerField(choices=UserType.choices, default=UserType.Free)
    is_mail_notification_allowed = models.BooleanField(default=True)
    is_app_notification_allowed = models.BooleanField(default=True)
    is_oauth_user = models.BooleanField(default=False)
    oauth_type = models.IntegerField(choices=OAuthType.choices, default=None, null=True)
    is_premium = models.BooleanField(default=False)
    is_staff = models.BooleanField(default=False)
    is_active = models.BooleanField(default=True)
    created_at = models.DateTimeField(auto_now_add=True, verbose_name="created at")
    updated_at = models.DateTimeField(auto_now=True, verbose_name="updated at")
    objects = UserManager()

    USERNAME_FIELD = 'email'
    REQUIRED_FIELDS = ['username']

    def __str__(self):
        return str(self.id) + ' ' + self.email + ' ' + self.username
