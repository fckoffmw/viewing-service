# w2g (watch together)

сервис для совместного просмотра видео с источником из вконтакте с чатом 

## функциональность

- плеер с источником из вк видео встроенный тегом iframe
<iframe src="https://vkvideo.ru/video_ext.php?oid=-231263435&id=456240185&hash=d1213ab8896c93b2&hd=4" width="1920" height="1080" allow="autoplay; encrypted-media; fullscreen; picture-in-picture; screen-wake-lock;" frameborder="0" allowfullscreen></iframe>

- простой чат с правой стороны для двоих на Go

## требования
- наличие регистрации / авторизации (session based) 
- отображение профилей при просмотре (аватар + имя, вместо me, notme)
- наличие страницы выбора контента (продумать категории)
- наличие страницы профиля, отображающей:
    - ник пользователя
    - аватарка
    - смотреть позже (названия текстом)
    - заметки о просмотренных видео
    - история просмотра*

## стек
- Go для чата
- простой клиент (html, css, js)
- развертывание несколькими командами на личном vps
- docker (docker-compose при необходимости)
- csv as storage
- redis for sessions

## реализация

storage (csv):
- getUserById(id) User

user model
- id
- username
- pass-hash


