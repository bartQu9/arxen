<template>
    <div>
        <div class="card-date" v-if="showDate" v-html="message.timeStamp"></div>
        <div
                class="message-box"
                :class="{ 'offset-current': message.user === currentUserId }"
        >
            <div
                    class="message-container"
                    :class="{
					'message-container-offset': messageOffset
				}"
            >
                <div
                        v-if="message.chatId"
                        class="text-username"
                        :class="{
							'username-reply': false
						}"
                >
                    <span v-html="message.user"></span>
                </div>
                <span v-html="message.text"></span>

                <div class="text-timestamp">
                    <span v-html="message.timeStamp"></span>
                </div>
            </div>
        </div>

    </div>
</template>

<script>
    export default {
        name: "Message",
        // tmp solution
        data() {
            return {
                currentUserId: "tcp://10.6.0.3:7878"
            }
        },
        props: {
            showDate: {type: Boolean, default: false },
            message: {
                type: Object,
            },
        },
        computed: {
            messageOffset() {
                return (
                    this.index > 0 &&
                    this.message.userId !== this.messages[this.index - 1].userId
                )
            },
        }
    };
    //<strong>{{message.user}}</strong>: {{message.text}}
</script>

<style scoped>

</style>
