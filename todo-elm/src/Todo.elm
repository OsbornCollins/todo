
module Todo exposing (main)
import Browser
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (onInput)



-- MAIN
main =
    Browser.sandbox { init = init, update = update, view = view }

-- MODEL

type alias Model =
  { task_name : String
  , description : String
  , notes : String
  , category : String
  , priority : String
  , status : String
  }


init : Model
init =
  Model "" "" "" "" "" ""



-- UPDATE


type Msg
  = Task_Name String
  | Description String
  | Notes String
  | Category String
  | Priority String
  | Status String


update : Msg -> Model -> Model
update msg model =
  case msg of
    Task_Name task_name ->
      { model | task_name = task_name }

    Description description ->
      { model | description = description }

    Notes notes ->
      { model | notes = notes }

    Category category ->
      { model | category = category }

    Priority priority ->
      { model | priority = priority }

    Status status ->
      { model | status = status }



-- VIEW


view : Model -> Html Msg
view model =
  div [ class "main" ] [
    div [ class "signup" ]
    [ Html.form [ action "http://localhost:4000/v1/todoitems", id "userform", method "POST" ]
        [ label [ attribute "aria-hidden" "true", for "chk" ]
            [ text "To-Do List Form" ]
        , div []
        [ viewInput "text" "Task Name" model.task_name Task_Name
        , viewInput "text" "Description" model.description Description
        , viewInput "text" "Notes" model.notes Notes
        , viewInput "text" "Category" model.category Category
        , viewInput "text" "Priority" model.priority Priority
        , viewInput "text" "Status" model.status Status
        , viewValidation model
        ]
        , button []
            [ text "Submit" ]
        ]
    ]
  ]


viewInput : String -> String -> String -> (String -> msg) -> Html msg
viewInput t p v toMsg =
  input [ type_ t, placeholder p, value v, onInput toMsg ] []


viewValidation : Model -> Html msg
viewValidation model =
  if model.task_name == "" || model.description == "" || model.notes == "" || model.category == "" || model.priority == "" || model.status == "" then
    div [ style "color" "red", style "text-align" "center" ] [ text "Please Fill All Fields!" ]
  else
    div [ style "color" "green",  style "text-align" "center" ] [ text "Good!" ]