# Generated by Django 4.1.7 on 2023-03-02 12:40

import core.models
from django.db import migrations, models


class Migration(migrations.Migration):

    initial = True

    dependencies = [
        ('auth', '0012_alter_user_first_name_max_length'),
    ]

    operations = [
        migrations.CreateModel(
            name='User',
            fields=[
                ('password', models.CharField(max_length=128, verbose_name='password')),
                ('last_login', models.DateTimeField(blank=True, null=True, verbose_name='last login')),
                ('is_superuser', models.BooleanField(default=False, help_text='Designates that this user has all permissions without explicitly assigning them.', verbose_name='superuser status')),
                ('id', models.AutoField(primary_key=True, serialize=False)),
                ('email', models.EmailField(max_length=255, unique=True)),
                ('username', models.CharField(max_length=255, unique=True)),
                ('image', models.ImageField(null=True, upload_to=core.models.upload_location)),
                ('fcm_token', models.CharField(default='', max_length=255)),
                ('refresh_token', models.CharField(max_length=255, null=True)),
                ('user_type', models.IntegerField(choices=[(1, 'Free'), (2, 'Premium')], default=1)),
                ('is_mail_notification_allowed', models.BooleanField(default=True)),
                ('is_app_notification_allowed', models.BooleanField(default=True)),
                ('is_oauth_user', models.BooleanField(default=False)),
                ('oauth_type', models.IntegerField(choices=[(1, 'Google')], default=None)),
                ('is_premium', models.BooleanField(default=False)),
                ('is_staff', models.BooleanField(default=False)),
                ('is_active', models.BooleanField(default=True)),
                ('created_at', models.DateTimeField(auto_now_add=True, verbose_name='created at')),
                ('updated_at', models.DateTimeField(auto_now=True, verbose_name='updated at')),
                ('groups', models.ManyToManyField(blank=True, help_text='The groups this user belongs to. A user will get all permissions granted to each of their groups.', related_name='user_set', related_query_name='user', to='auth.group', verbose_name='groups')),
                ('user_permissions', models.ManyToManyField(blank=True, help_text='Specific permissions for this user.', related_name='user_set', related_query_name='user', to='auth.permission', verbose_name='user permissions')),
            ],
            options={
                'abstract': False,
            },
        ),
    ]
