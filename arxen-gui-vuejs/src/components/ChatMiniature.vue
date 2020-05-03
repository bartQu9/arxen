<template>
    <div>
        <strong>{{chat.chatId}}</strong>
    </div>
</template>

<script>
    export default {
        name: "ChatMiniature",
        props: {
            chat: {
                type: Object,
            },
        },
    }
</script>

<style scoped>

</style>
