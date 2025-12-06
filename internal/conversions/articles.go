package conversions

import (
	"errors"
	"fmt"

	"code.superseriousbusiness.org/activity/streams"
	"code.superseriousbusiness.org/activity/streams/vocab"
	"github.com/sidereusnuntius/gowiki/internal/domain"
	"github.com/sidereusnuntius/gowiki/internal/federation"
)

func ConvertArticle(article vocab.Type) (result domain.ArticleFed, err error) {
	fmt.Println("Type:", streams.ActivityStreamsArticleName)
	switch article.GetTypeName() {
	case streams.ActivityStreamsArticleName:
		a, ok := article.(vocab.ActivityStreamsArticle)
		if !ok {
			err = fmt.Errorf("conversion to Article failed")
			return
		}
		result, err = convertArticle(a)
	default:
		err = fmt.Errorf("%w: %s", errors.ErrUnsupported, article.GetTypeName())
	}

	return
}

func convertArticle(article vocab.ActivityStreamsArticle) (domain.ArticleFed, error) {
	result := domain.ArticleFed{}
	id := article.GetJSONLDId()
	if id == nil {
		return result, fmt.Errorf("%w: id", federation.ErrMissingProperty)
	}
	result.ApID = id.Get()

	// TODO: handle multilingual titles, summaries, sources etc.
	title := article.GetActivityStreamsName()
	if title == nil || title.Len() == 0 {
		return domain.ArticleFed{}, fmt.Errorf("%w: name", federation.ErrMissingProperty)
	}
	result.Title = title.Begin().GetXMLSchemaString()

	if summary := article.GetActivityStreamsSummary(); summary != nil && summary.Len() != 0 {
		result.Summary = summary.Begin().GetXMLSchemaString()
	}

	var content, mediaType string
	var err error
	if source := article.GetActivityStreamsSource(); source != nil {
		content, mediaType, err = processSourceProperty(source)
	} else if contentProp := article.GetActivityStreamsContent(); contentProp != nil && contentProp.Len() != 0 {
		content = contentProp.Begin().GetXMLSchemaString()
		if mediaTypeProp := article.GetActivityStreamsMediaType(); mediaTypeProp != nil {
			mediaType = mediaTypeProp.Get()
		}
	} else {
		err = fmt.Errorf("%w: source or content properties", federation.ErrMissingProperty)
	}
	if err != nil {
		return domain.ArticleFed{}, err
	}

	result.Content = content
	result.MediaType = mediaType
	result.Host = result.ApID.Host

	if attributedTo := article.GetActivityStreamsAttributedTo(); attributedTo != nil && attributedTo.Len() != 0 {
		result.AttributedTo = attributedTo.Begin().GetIRI()
	}

	if url := article.GetActivityStreamsUrl(); url != nil && url.Len() != 0 {
		result.Url = url.Begin().GetIRI()
	}
	// TODO: protected, license,

	if published := article.GetActivityStreamsPublished(); published != nil {
		result.Published = published.Get()
	}

	if updated := article.GetActivityStreamsUpdated(); updated != nil {
		result.LastUpdated = updated.Get()
	}

	return result, nil
}

func processSourceProperty(prop vocab.ActivityStreamsSourceProperty) (source, mediaType string, err error) {
	// TODO: handle case with multiple languages.
	if prop.IsIRI() {
		err = fmt.Errorf("%w: source is an IRI", federation.ErrUnprocessablePropValue)
		return
	}
	serial, err := prop.Serialize()
	if err != nil {
		return
	}
	sourceMap, ok := serial.(map[string]any)
	if !ok {
		err = fmt.Errorf("%w: failed to convert into map", federation.ErrUnprocessablePropValue)
		return
	}

	source, ok = sourceMap["content"].(string)
	if !ok {
		err = errors.Join(err, fmt.Errorf("%w: source content", federation.ErrMissingProperty))
		return
	}

	mediaType, _ = sourceMap["mediaType"].(string)

	return
}
