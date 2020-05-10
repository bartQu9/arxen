<template>

    <div ref="roomFooter" class="room-footer">
        <div class="box-footer"></div>
        <textarea
                :placeholder="textMessages.TYPE_MESSAGE"
                v-model="messageInput"
                @input="onChangeInput"
        ></textarea>

        <div class="icon-textarea">
            <div
                    @click="onPostClick"
                    class="svg-button"
                    :class="{ 'send-disabled': inputDisabled }"
            >
                <svg-icon name="send" :param="inputDisabled ? 'disabled' : ''" />
            </div>

        </div>


    </div>
</template>

<script>
    import gql from 'graphql-tag';
    import SvgIcon from "@/components/SvgIcon";

    export default {
        name: "MessageForm",
        components: {
            'svg-icon': SvgIcon,
        },
        data() {
            return {
                messageInput: '',
            };
        },
        props: {
            textMessages: { type: Object, required: true },
            selectedChatId: {type: String}
        },
        computed: {
            inputDisabled() {
                return this.isMessageEmpty()
            },
        },
        methods: {
            onPostClick() {
                const messageInput = this.messageInput;
                this.$apollo
                    .mutate({
                        mutation: gql`mutation($chatID: String!, $text: String!) {postMessage(chatID: $chatID, text: $text) { chatId user text timeStamp }}`,
                        variables: {
                            chatID: this.selectedChatId,
                            text: messageInput,
                        },
                    })
                    .then(() => {
                        this.messageInput = '';
                    })
                    .catch((e) => {
                        console.error(e);
                    });
            },
            resizeTextarea() {
                const el = this.$refs['roomTextarea'];
                const padding = window
                    .getComputedStyle(el, null)
                    .getPropertyValue('padding-top')
                    .replace('px', '');
                el.style.height = 0;
                el.style.height = el.scrollHeight - padding * 2 + 'px'
            },
            isMessageEmpty() {
                return !this.messageInput.trim()
            },
            onChangeInput() {
                this.resizeTextarea();
                //this.$emit('typingMessage', this.message)
            },
        },
    };


    // <form class="col-12"
    // v-on:submit.prevent="onPostClick">
    //
    //     <div class="input-group">
    //     <input type="text"
    // class="form-control"
    // placeholder="Message..."
    // v-model.trim="messageInput">
    //     <div class="input-group-append">
    //     <button class="btn btn-outline-secondary"
    // type="submit">Post
    //     </button>
    //     </div>
    //     </div>
    //
    //     </form>
</script>

<style lang="scss" scoped>
    .col-messages {
        position: relative;
        height: 100%;
        flex: 1;
        overflow: hidden;
        display: flex;
        flex-flow: column;
    }
    .room-header {
        position: absolute;
        display: flex;
        align-items: center;
        height: 64px;
        width: 100%;
        z-index: 10;
        margin-right: 1px;
        background: var(--chat-header-bg-color);
        border-top-right-radius: var(--chat-container-border-radius);
    }
    .room-wrapper {
        display: flex;
        align-items: center;
        width: 100%;
        padding: 0 16px;
    }
    .toggle-button {
        margin-right: 15px;
        svg {
            height: 26px;
            width: 26px;
        }
    }
    .rotate-icon {
        transform: rotate(180deg) !important;
    }
    .room-name {
        font-size: 17px;
        font-weight: 500;
        line-height: 22px;
        color: var(--chat-header-color-name);
    }
    .room-info {
        font-size: 13px;
        line-height: 18px;
        color: var(--chat-header-color-info);
    }
    .room-options {
        margin-left: auto;
    }
    .container-scroll {
        background: var(--chat-content-bg-color);
        flex: 1;
        overflow-y: scroll;
        margin-right: 1px;
        margin-top: 60px;
        -webkit-overflow-scrolling: touch;
    }
    .messages-container {
        padding: 0 5px 5px;
    }
    .text-started {
        font-size: 14px;
        color: var(--chat-message-color-started);
        font-style: italic;
        text-align: center;
        margin-top: 30px;
        margin-bottom: 20px;
    }
    .icon-scroll {
        position: absolute;
        bottom: 80px;
        right: 20px;
        padding: 8px;
        background: var(--chat-bg-scroll-icon);
        border-radius: 50%;
        box-shadow: 0 1px 1px -1px rgba(0, 0, 0, 0.2), 0 1px 1px 0 rgba(0, 0, 0, 0.14),
        0 1px 2px 0 rgba(0, 0, 0, 0.12);
        display: flex;
        cursor: pointer;
        svg {
            height: 25px;
            width: 25px;
        }
    }
    .room-footer {
        width: calc(100% - 1px);
        border-bottom-right-radius: 4px;
        z-index: 10;
    }
    .box-footer {
        display: flex;
        position: relative;
        background: var(--chat-footer-bg-color);
        padding: 10px 8px 10px;
    }
    .reply-container {
        display: flex;
        padding: 10px 10px 0 10px;
        background: var(--chat-content-bg-color);
        align-items: center;
        max-width: 100%;
        .reply-box {
            width: 100%;
            overflow: hidden;
            background: var(--chat-footer-bg-color-reply);
            border-radius: 4px;
            padding: 8px 10px;
            display: flex;
        }
        .reply-info {
            overflow: hidden;
        }
        .reply-username {
            color: var(--chat-message-color-reply-username);
            font-size: 12px;
            line-height: 15px;
            margin-bottom: 2px;
        }
        .reply-content {
            font-size: 12px;
            color: var(--chat-message-color-reply-content);
        }
        .icon-reply {
            margin-left: 10px;
            svg {
                height: 20px;
                width: 20px;
            }
        }
        .image-reply {
            max-height: 100px;
            margin-right: 10px;
        }
    }
    textarea {
        height: 20px;
        width: 100%;
        line-height: 20px;
        overflow: hidden;
        outline: 0;
        resize: none;
        border-radius: 20px;
        padding: 12px 16px;
        box-sizing: content-box;
        font-size: 16px;
        background: var(--chat-bg-color-input);
        color: var(--chat-color);
        caret-color: var(--chat-color-caret);
        border: var(--chat-border-style-input);
        &::placeholder {
            color: var(--chat-color-placeholder);
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }
    }
    .textarea-outline {
        border: 1px solid var(--chat-border-color-input-selected);
        box-shadow: inset 0px 0px 0px 1px var(--chat-border-color-input-selected);
    }
    .icon-textarea {
        display: flex;
        margin: 12px 0 0 5px;
        svg,
        .wrapper {
            margin: 0 7px;
        }
    }
    .image-container {
        position: absolute;
        max-width: 25%;
        left: 16px;
        top: 18px;
    }
    .image-file {
        display: flex;
        justify-content: center;
        flex-direction: column;
        min-height: 30px;
        img {
            border-radius: 15px;
            width: 100%;
            max-width: 150px;
            max-height: 100%;
        }
    }
    .icon-image {
        position: absolute;
        top: 6px;
        left: 6px;
        z-index: 10;
        svg {
            height: 20px;
            width: 20px;
            border-radius: 50%;
        }
        &:before {
            content: ' ';
            position: absolute;
            width: 100%;
            height: 100%;
            background: rgba(0, 0, 0, 0.5);
            border-radius: 50%;
            z-index: -1;
        }
    }
    .file-container {
        display: flex;
        align-items: center;
        width: calc(100% - 75px);
        height: 20px;
        padding: 12px 0;
        box-sizing: content-box;
        background: var(--chat-bg-color-input);
        border: var(--chat-border-style-input);
        border-radius: 20px;
    }
    .file-container-edit {
        width: calc(100% - 109px);
    }
    .file-message {
        max-width: calc(100% - 75px);
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
    }
    .icon-file {
        display: flex;
        margin: 0 8px 0 15px;
    }
    .icon-remove {
        margin-left: 8px;
        svg {
            height: 18px;
            width: 18px;
        }
    }
    .send-disabled,
    .send-disabled svg {
        cursor: none !important;
        pointer-events: none !important;
        transform: none !important;
    }
    .messages-hidden {
        opacity: 0;
    }
    @media only screen and (max-width: 768px) {
        .room-header {
            height: 50px;
            .room-wrapper {
                padding: 0 10px;
            }
            .room-name {
                font-size: 16px;
                line-height: 22px;
            }
            .room-info {
                font-size: 12px;
                line-height: 16px;
            }
            .room-avatar {
                height: 37px;
                width: 37px;
            }
        }
        .container-scroll {
            margin-top: 50px;
        }
        .box-footer {
            border-top: var(--chat-border-style-input);
            padding: 7px 2px 7px 7px;
        }
        .text-started {
            margin-top: 20px;
        }
        textarea {
            padding: 7px;
            line-height: 18px;
            &::placeholder {
                color: transparent;
            }
        }
        .icon-textarea {
            margin: 6px 0 0 5px;
            svg,
            .wrapper {
                margin: 0 5px;
            }
        }
        .image-container {
            top: 10px;
            left: 10px;
        }
        .image-file img {
            transform: scale(0.97);
        }
        .room-footer {
            width: 100%;
        }
        .file-container {
            padding: 7px 0;
            .icon-file {
                margin-left: 10px;
            }
        }
        .reply-container {
            padding: 5px 8px;
        }
        .icon-scroll {
            bottom: 70px;
        }
    }
</style>
