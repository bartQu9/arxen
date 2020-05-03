<template>
    <div>
        <strong>{{message.user}}</strong>: {{message.text}}
    </div>
</template>

<script>
    export default {
        name: "Message",
        props: {
            message: {
                type: Object,
            },
        },
    };
</script>

<style scoped>

</style>
